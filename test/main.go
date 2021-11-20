package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"math/rand"
	"nir/test/contract"
	"nir/test/entity"
	"nir/test/entity/account"
	"nir/test/exchange"
	"nir/test/startbalance"
	"nir/test/user"
	"nir/test/writer"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	ctx, closeDeps := prepareDeps(context.Background())
	defer closeDeps()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Println("Creating entities...")
	users, exchanges, err := createEntitiesWithEther(ctx)

	log.Println("Start airdrop...")

	for i := 0; i < countTokens; i++ {
		err = airdrop(ctx, users)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Start sending transactions to exchange...")

	sendTransactionsToExchange(ctx, users, exchanges, countTransactions)

	log.Println("Closing exchanges...")

	closeExchanges(exchanges)

	log.Println("Saving blockchain...")
	err = saveBlockchain(ctx, exchanges)
	if err != nil {
		log.Fatal(err)
	}
}

func addEtherToEntities(ctx context.Context, exchanges []*exchange.Exchange, users []*user.User) error {
	var (
		entities []entity.Entity
	)

	for _, u := range users {
		err := CreateExchangesAccounts(exchanges, u)
		if err != nil {
			return err
		}

		entities = append(entities, u)
	}

	for _, e := range exchanges {
		entities = append(entities, e)
	}

	entity.AddEtherToEntities(ctx, entities, commonAmount)

	return nil
}

func createEOAs(count, size uint64) (clusters []*user.User, err error) {
	clusters = make([]*user.User, count)

	for i := uint64(0); i < count; i++ {
		clusters[i], err = user.CreateUserFromSize(size)
		if err != nil {
			return nil, err
		}
	}

	return clusters, nil
}

func createExchanges(ctx context.Context, count uint64) (exchanges []*exchange.Exchange, err error) {
	exchanges = make([]*exchange.Exchange, count)

	for i := uint64(0); i < count; i++ {
		exchanges[i], err = exchange.CreateExchange(ctx)
		if err != nil {
			return nil, err
		}
	}

	return exchanges, nil
}

func CreateExchangesAccounts(exchanges []*exchange.Exchange, entity *user.User) error {
	for _, exch := range exchanges {
		err := entity.CreateExchangeAccounts(exch)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendTransactionsToExchange(ctx context.Context, entities []*user.User, exchanges []*exchange.Exchange, countTxs int32) {
	var (
		count int32
		wg    sync.WaitGroup
	)

	txsNumbers := new(int32)

	for {
		for j := 0; j < transactonsPerSecond; j++ {
			count++

			wg.Add(1)

			go func() {
				defer wg.Done()

				amount := rand.Intn(100)

				currentTxNumber := atomic.AddInt32(txsNumbers, 1)

				currentEntity := entities[currentTxNumber%int32(len(entities))]
				currentExchange := exchanges[currentTxNumber%int32(len(exchanges))]

				tx, err := currentEntity.SendTransaction(ctx, currentExchange.GetName(), int64(amount))
				if err != nil {
					log.Printf("can't send %d tranasction: %v\n", currentTxNumber, err)

					return
				}

				currentExchange.AddIncomingTransaction(tx)
			}()
			if count == countTxs {
				wg.Wait()

				return
			}
		}

		log.Println("Transactions sent", atomic.LoadInt32(txsNumbers))
		time.Sleep(time.Second)
	}
}

func closeExchanges(exchanges []*exchange.Exchange) {
	var wg sync.WaitGroup

	for _, exch := range exchanges {
		wg.Add(1)

		exch := exch
		go func() {
			defer wg.Done()

			exch.Close()
		}()
	}

	wg.Wait()
}

func addEthToAccount(ctx context.Context, acc *common.Address, amount int64) error {
	waitCh := make(chan struct{})

	err := startbalance.AddTask(ctx, func(a *account.Account) {
		_, err := a.SendTransaction(ctx, acc, amount, true)
		if err != nil {
			log.Fatal(err)
		}

		close(waitCh)
	})
	if err != nil {
		return err
	}

	<-waitCh

	return nil
}

func createUsers() []*user.User {
	clustersDepositReuse, err := createEOAs(countCluster, maxCountAccountsInCluster)
	if err != nil {
		log.Fatal(err)
	}

	for _, cluster := range clustersDepositReuse {
		fmt.Println("CLUSTER ACCOUNTS", cluster.GetAccounts())
	}

	accounts, err := createEOAs(countAccounts-(countCluster*maxCountAccountsInCluster), 1)
	if err != nil {
		log.Fatal(err)
	}

	return append(accounts, clustersDepositReuse...)
}

func prepareDeps(ctx context.Context) (context.Context, func()) {
	var (
		wr  *writer.Writer
		err error
	)

	for {
		wr, err = writer.Connect(ctx, getWriterNodes())
		if err != nil {
			continue
		}

		break
	}

	ctx = writer.WithWriter(ctx, wr)
	ctx, err = startbalance.CommonBalancesWithCtx(ctx, getAccountsWithBalance())
	if err != nil {
		log.Fatal(err)
	}

	return ctx, func() {
		startbalance.Close(ctx)
	}
}

func deployContract(ctx context.Context, distributor *account.Account, tokens int) (*contract.SimpleToken, error) {
	bytecode, err := contract.BytecodeContract(contract.SimpleTokenBin, contract.SimpleTokenABI, "test", "t", big.NewInt(int64(tokens)))
	if err != nil {
		return nil, err
	}

	var gas uint64

	err = writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		gas, innerErr = w.EstimateGas(ctx, *distributor.GetAddress(), nil, bytecode)

		return innerErr
	})
	if err != nil {
		return nil, err
	}

	err = writer.Execute(ctx, func(w *writer.Writer) error {
		b, innerErr := w.BalanceAt(ctx, *distributor.GetAddress())
		if innerErr != nil {
			return innerErr
		}

		fmt.Println("Distributor balance before deploy", b.Int64())

		return nil
	})
	if err != nil {
		return nil, err
	}

	var token *contract.SimpleToken

	tx, err := distributor.ExecuteContract(ctx, gas, func(auth *bind.TransactOpts, backend bind.ContractBackend) (tx *types.Transaction, innerErr error) {
		_, tx, token, innerErr = contract.DeploySimpleToken(auth, backend, "TestContract", "TC", big.NewInt(100000000000))

		return tx, innerErr
	})
	if err != nil {
		return nil, err
	}

	err = writer.Execute(ctx, func(w *writer.Writer) error {
		return w.WaitTx(ctx, tx.Hash())
	})

	err = writer.Execute(ctx, func(w *writer.Writer) error {
		b, innerErr := w.BalanceAt(ctx, *distributor.GetAddress())
		if innerErr != nil {
			return innerErr
		}

		fmt.Println("Distributor balance after deploy", b.Int64(), "tx hash", tx.Hash().String())

		return nil
	})
	if err != nil {
		return nil, err
	}

	return token, err
}

func airdrop(ctx context.Context, users []*user.User) error {
	distributor, err := account.CreateAccount()
	if err != nil {
		return err
	}

	err = addEthToAccount(ctx, distributor.GetAddress(), commonAmount)
	if err != nil {
		return err
	}

	tokenContract, err := deployContract(ctx, distributor, countAccounts)
	if err != nil {
		return err
	}

	err = transferTokensToUsers(ctx, users, tokenContract, distributor, big.NewInt(1))
	if err != nil {
		return err
	}

	for _, u := range users {
		err := u.CollectTokenOnOneAcc(ctx, mixTokensDepth, func(address *common.Address) (*big.Int, error) {
			return tokenContract.BalanceOf(nil, *address)
		}, func(ctx context.Context, from *account.Account, toAddr *common.Address, amount *big.Int) (innerErr error) {
			return transferToken(ctx, tokenContract, from, toAddr, amount)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func transferTokensToUsers(ctx context.Context, users []*user.User, tokenContract *contract.SimpleToken, distributor *account.Account, tokens *big.Int) error {
	for i := 0; i < len(users); i += 4 {
		randIdx := i + rand.Intn(4)
		if randIdx > len(users)-1 {
			return nil
		}

		for _, acc := range users[randIdx].GetAccounts() {
			err := transferToken(ctx, tokenContract, distributor, acc, tokens)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func transferToken(ctx context.Context, tokenContract *contract.SimpleToken, distributor *account.Account, toAddr *common.Address, tokens *big.Int) error {
	gas, err := contract.EstimateTransfer(ctx, *distributor.GetAddress(), toAddr, tokens)
	if err != nil {
		return err
	}

	tx, err := distributor.ExecuteContract(ctx, gas, func(auth *bind.TransactOpts, backend bind.ContractBackend) (*types.Transaction, error) {
		return tokenContract.Transfer(auth, *toAddr, tokens)
	})
	if err != nil {
		return err
	}

	return writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		return w.WaitTx(ctx, tx.Hash())
	})
}

func saveBlockchain(ctx context.Context, exchanges []*exchange.Exchange) error {
	var currentBlock uint64

	err := writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		currentBlock, innerErr = w.BlockNumber(ctx)
		return innerErr
	})
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "./test/download_chain/download.sh", strconv.FormatUint(currentBlock, 10))
	if err != nil {
		return err
	}

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("%w: stderr = %q, stdout = %q", err, stderr.String(), stdout.String())
	}

	return exchange.SaveExchanges(exchanges)
}

func createEntitiesWithEther(ctx context.Context) ([]*user.User, []*exchange.Exchange, error) {
	users := createUsers()

	exchanges, err := createExchanges(ctx, countExchanges)
	if err != nil {
		return nil, nil, err
	}

	err = addEtherToEntities(ctx, exchanges, users)
	if err != nil {
		return nil, nil, err
	}

	return users, exchanges, nil
}
