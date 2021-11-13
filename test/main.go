package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
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
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	defer startbalance.Close(ctx)

	exchanges, err := createExchanges(ctx, countExchanges)
	if err != nil {
		log.Fatal(err)
	}

	clusters, err := createEOAs(countCluster, maxCountAccountsInCluster)
	if err != nil {
		log.Fatal(err)
	}

	for _, cluster := range clusters {
		fmt.Println("CLUSTER ACCOUNTS", cluster.GetAccounts())
	}

	accounts, err := createEOAs(countAccounts-(countCluster*maxCountAccountsInCluster), 1)
	if err != nil {
		log.Fatal(err)
	}

	users := append(accounts, clusters...)

	err = addEtherToEntities(ctx, exchanges, users)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Deploy contract...")
	acc, err := account.CreateAccount()
	if err != nil {
		log.Fatal(err)
	}

	err = addEthToAccount(ctx, acc.GetAddress(), commonAmount)
	if err != nil {
		log.Fatal(err)
	}

	balance, err := wr.BalanceAt(ctx, *acc.GetAddress())
	if err != nil {
		log.Fatal(err)
	}

	bytecode, err := bytecodeContract(contract.SimpleTokenBin, contract.SimpleTokenABI, "test", "t", big.NewInt(1000))
	if err != nil {
		log.Fatal(err)
	}

	estimate, err := wr.EstimateGas(ctx, *acc.GetAddress(), nil, bytecode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DEPLOYER BALANCE", balance, estimate)

	var token *contract.SimpleToken

	err = acc.DeployContract(ctx, estimate, func(auth *bind.TransactOpts, backend bind.ContractBackend) (innerErr error) {
		var tx *types.Transaction

		_, tx, token, innerErr = contract.DeploySimpleToken(auth, backend, "TestContract", "TC", big.NewInt(100))
		if innerErr != nil {
			fmt.Println("CONTRACT ERR", innerErr)

			return innerErr
		}

		err = wr.WaitTx(ctx, tx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	rndAcc, err := account.CreateAccount()
	if err != nil {
		log.Fatal(err)
	}

	err = acc.ExecuteContract(ctx, estimate, func(auth *bind.TransactOpts, backend bind.ContractBackend) error {
		tx, innerErr := token.Transfer(auth, *rndAcc.GetAddress(), big.NewInt(100))
		if innerErr != nil {
			return innerErr
		}

		err = wr.WaitTx(ctx, tx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		b, innerErr := token.BalanceOf(nil, *rndAcc.GetAddress())
		if innerErr != nil {
			return innerErr
		}

		fmt.Println("BALANCE OF NEW ACC", b.Int64())

		return nil
	})

	fmt.Println("Start sending transactions...")

	SendTransactions(ctx, users, exchanges, countTransactions)

	fmt.Println("Closing exchanges...")

	closeExchanges(exchanges)

	var currentBlock uint64

	err = writer.Execute(ctx, func(w *writer.Writer) (innerErr error) {
		currentBlock, innerErr = w.BlockNumber(ctx)
		return innerErr
	})
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("bash", "./test/download_chain/download.sh", strconv.FormatUint(currentBlock, 10))
	if err != nil {
		log.Fatal(err)
	}

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err = cmd.Run(); err != nil {
		log.Fatal(err, "output", stderr.String())
	}

	err = exchange.SaveExchanges(exchanges)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successful\n", stdout.String())
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

func SendTransactions(ctx context.Context, entities []*user.User, exchanges []*exchange.Exchange, countTxs int32) {
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

		fmt.Println("Transactions sent", atomic.LoadInt32(txsNumbers))
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

func bytecodeContract(bin, abiJSON string, args ...interface{}) ([]byte, error) {
	bytecode := common.FromHex(bin)

	ssAbi, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}

	ssInput, err := ssAbi.Pack("", args...)
	if err != nil {
		return nil, err
	}

	return append(bytecode, ssInput...), nil
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
