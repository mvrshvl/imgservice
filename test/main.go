package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"nir/test/entity"
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

	accounts, err := createEOAs(countAccounts-(countCluster*maxCountAccountsInCluster), 1)
	if err != nil {
		log.Fatal(err)
	}

	users := append(accounts, clusters...)

	err = addEtherToEntities(ctx, exchanges, users)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Start sending transactions...")

	SendTransactions(ctx, users, exchanges, countTransactions)

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
		clusters[i], err = user.CreateEntity(size)
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
