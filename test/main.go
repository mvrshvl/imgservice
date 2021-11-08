package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"nir/test/entity"
	"nir/test/exchange"
	"nir/test/user"
	"nir/test/writer"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	ctx := context.Background()

	var (
		wr  *ethclient.Client
		err error
	)

	for {
		wr, err = writer.Connect(ctx, writerNode)
		if err != nil {
			continue
		}

		break
	}

	ctx = writer.WithWriter(ctx, wr)

	exchanges, err := createExchanges(countExchanges)
	if err != nil {
		log.Fatal(err)
	}

	clusters, err := createEOAs(countCluster, maxCountAccountsInCluster)
	if err != nil {
		log.Fatal(err)
	}

	accounts, err := createEOAs(countAccounts-(countCluster*maxCountAccountsInCluster), 1)
	if err != nil {
		log.Fatal(err)
	}

	users := append(accounts, clusters...)

	addEtherToEOAs(ctx, exchanges, users)

	SendTransactions(ctx, users, exchanges, countTransactions)

	currentBlock, err := wr.BlockNumber(ctx)
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

	fmt.Println("Successful\n", stdout.String())
}

func addEtherToEOAs(ctx context.Context, exchanges []*exchange.Exchange, eoas []*user.User) {
	var (
		wg sync.WaitGroup
	)

	for _, acc := range eoas {
		err := CreateExchangesAccounts(exchanges, acc)
		if err != nil {
			log.Fatal(err)
		}

		var ent entity.Entity

		ent = acc

		wg.Add(1)

		go func() {
			defer wg.Done()

			err := entity.AddEtherToEntity(ctx, ent, commonAmount)
			if err != nil {
				log.Fatal(err)
			}
		}()

		time.Sleep(time.Millisecond * 50)
	}

	wg.Wait()
}

func createEOAs(count, size uint64) (clusters []*user.User, err error) {
	clusters = make([]*user.User, count)

	for i := uint64(0); i < count; i++ {
		clusters[i], err = user.CreateEntity(size)
		if err != nil {
			return nil, err
		}
	}

	return clusters, nil
}

func createExchanges(count uint64) (exchanges []*exchange.Exchange, err error) {
	exchanges = make([]*exchange.Exchange, count)

	for i := uint64(0); i < count; i++ {
		exchanges[i], err = exchange.CreateExchange()
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

				currentTxNumber := atomic.AddInt32(txsNumbers, 1)

				currentEntity := entities[currentTxNumber%int32(len(entities))]
				currentExchange := exchanges[currentTxNumber%int32(len(exchanges))]

				deposit, err := currentEntity.SendTransaction(ctx, currentExchange.GetName(), 1)
				if err != nil {
					log.Printf("can't send %d tranasction: %v\n", currentTxNumber, err)

					return
				}

				err = currentExchange.GetEthFromDeposit(ctx, deposit, 1)
				if err != nil {
					log.Printf("can't send %d tranasction from deposit: %v\n", currentTxNumber, err)

					return
				}
			}()
			if count == countTxs {
				wg.Wait()

				return
			}
		}

		time.Sleep(time.Second)
	}
}
