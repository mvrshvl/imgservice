package contracts

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"log"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func Test_SendRawTx(t *testing.T) {
	client, err := ethclient.Dial("http://localhost:59047")
	if err != nil {
		t.Fatal(err)
	}

	to := common.HexToAddress("0x13e3cf8EDEb694952f90Cb791B31D7A970151614")

	ctx := context.Background()

	balanceBefore, err := client.BalanceAt(ctx, to, nil)
	if err != nil {
		t.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA("29f4455cd82770e305096f33c2a53f13efed2974873c883a6d7ca7d1bdcdf0c7")
	if err != nil {
		t.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &to,
		Value:    big.NewInt(100),
	})

	chainID := big.NewInt(45439)

	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(chainID), privateKey)

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		t.Fatal(err)
	}

	txReceipt, err := WaitTickerTx(ctx, client, signedTx.Hash())
	if err != nil {
		t.Fatal(err)
	}

	if txReceipt.Status == types.ReceiptStatusFailed {
		t.Fatal(err)
	}

	balanceAfter, err := client.BalanceAt(ctx, to, nil)
	if err != nil {
		t.Fatal(err)
	}

	if balanceBefore.Uint64()+100 != balanceAfter.Uint64() {
		t.Fatalf("Expected %v:, result %v", balanceBefore.Uint64()+100, balanceAfter.Uint64())
	}
}

func WaitTickerTx(ctx context.Context, client *ethclient.Client, hash common.Hash) (*types.Receipt, error) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for range tick.C {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout")
		default:
			txReceipt, err := client.TransactionReceipt(ctx, hash)
			if err != nil {
				if errors.Is(err, ethereum.NotFound) {
					continue
				}

				return nil, err
			}

			return txReceipt, nil
		}
	}

	return nil, nil
}

const (
	bytecode    = `{ "linkReferences": {}, "object": "60806040526801a055690d9db800003414610065576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260218152602001806106606021913960400191505060405180910390fd5b33600260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060008060006101000a81548160ff021916908360ff160217905550610590806100d06000396000f3fe6080604052600436106100555760003560e01c80630c11dedd1461005a578063138fbe71146100bf5780632e1a7d4d146100ea5780638da5cb5b14610139578063b69ef8a814610190578063d0e30db0146101bb575b600080fd5b34801561006657600080fd5b506100a96004803603602081101561007d57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506101d9565b6040518082815260200191505060405180910390f35b3480156100cb57600080fd5b506100d46102d2565b6040518082815260200191505060405180910390f35b3480156100f657600080fd5b506101236004803603602081101561010d57600080fd5b81019080803590602001909291905050506102f1565b6040518082815260200191505060405180910390f35b34801561014557600080fd5b5061014e610415565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561019c57600080fd5b506101a561043b565b6040518082815260200191505060405180910390f35b6101c3610482565b6040518082815260200191505060405180910390f35b6000803073ffffffffffffffffffffffffffffffffffffffff16311015801561024f5750600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16145b156102b3578173ffffffffffffffffffffffffffffffffffffffff166108fc3073ffffffffffffffffffffffffffffffffffffffff16319081150290604051600060405180830381858888f193505050501580156102b1573d6000803e3d6000fd5b505b3073ffffffffffffffffffffffffffffffffffffffff16319050919050565b60003073ffffffffffffffffffffffffffffffffffffffff1631905090565b6000600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205482116103ce5781600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825403925050819055503373ffffffffffffffffffffffffffffffffffffffff166108fc839081150290604051600060405180830381858888f193505050501580156103cc573d6000803e3d6000fd5b505b600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905090565b600034600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825401925050819055503373ffffffffffffffffffffffffffffffffffffffff167fa8126f7572bb1fdeae5b5aa9ec126438b91f658a07873f009d041ae690f3a193346040518082815260200191505060405180910390a2600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490509056fea165627a7a72305820753d1cc549a04ad30e75834d6c8210a508cf241a6b16b4d478a2f206a121608c0029333020657468657220696e697469616c2066756e64696e67207265717569726564", "opcodes": "PUSH1 0x80 PUSH1 0x40 MSTORE PUSH9 0x1A055690D9DB80000 CALLVALUE EQ PUSH2 0x65 JUMPI PUSH1 0x40 MLOAD PUSH32 0x8C379A000000000000000000000000000000000000000000000000000000000 DUP2 MSTORE PUSH1 0x4 ADD DUP1 DUP1 PUSH1 0x20 ADD DUP3 DUP2 SUB DUP3 MSTORE PUSH1 0x21 DUP2 MSTORE PUSH1 0x20 ADD DUP1 PUSH2 0x660 PUSH1 0x21 SWAP2 CODECOPY PUSH1 0x40 ADD SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 REVERT JUMPDEST CALLER PUSH1 0x2 PUSH1 0x0 PUSH2 0x100 EXP DUP2 SLOAD DUP2 PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF MUL NOT AND SWAP1 DUP4 PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND MUL OR SWAP1 SSTORE POP PUSH1 0x0 DUP1 PUSH1 0x0 PUSH2 0x100 EXP DUP2 SLOAD DUP2 PUSH1 0xFF MUL NOT AND SWAP1 DUP4 PUSH1 0xFF AND MUL OR SWAP1 SSTORE POP PUSH2 0x590 DUP1 PUSH2 0xD0 PUSH1 0x0 CODECOPY PUSH1 0x0 RETURN INVALID PUSH1 0x80 PUSH1 0x40 MSTORE PUSH1 0x4 CALLDATASIZE LT PUSH2 0x55 JUMPI PUSH1 0x0 CALLDATALOAD PUSH1 0xE0 SHR DUP1 PUSH4 0xC11DEDD EQ PUSH2 0x5A JUMPI DUP1 PUSH4 0x138FBE71 EQ PUSH2 0xBF JUMPI DUP1 PUSH4 0x2E1A7D4D EQ PUSH2 0xEA JUMPI DUP1 PUSH4 0x8DA5CB5B EQ PUSH2 0x139 JUMPI DUP1 PUSH4 0xB69EF8A8 EQ PUSH2 0x190 JUMPI DUP1 PUSH4 0xD0E30DB0 EQ PUSH2 0x1BB JUMPI JUMPDEST PUSH1 0x0 DUP1 REVERT JUMPDEST CALLVALUE DUP1 ISZERO PUSH2 0x66 JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH2 0xA9 PUSH1 0x4 DUP1 CALLDATASIZE SUB PUSH1 0x20 DUP2 LT ISZERO PUSH2 0x7D JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST DUP2 ADD SWAP1 DUP1 DUP1 CALLDATALOAD PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND SWAP1 PUSH1 0x20 ADD SWAP1 SWAP3 SWAP2 SWAP1 POP POP POP PUSH2 0x1D9 JUMP JUMPDEST PUSH1 0x40 MLOAD DUP1 DUP3 DUP2 MSTORE PUSH1 0x20 ADD SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 RETURN JUMPDEST CALLVALUE DUP1 ISZERO PUSH2 0xCB JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH2 0xD4 PUSH2 0x2D2 JUMP JUMPDEST PUSH1 0x40 MLOAD DUP1 DUP3 DUP2 MSTORE PUSH1 0x20 ADD SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 RETURN JUMPDEST CALLVALUE DUP1 ISZERO PUSH2 0xF6 JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH2 0x123 PUSH1 0x4 DUP1 CALLDATASIZE SUB PUSH1 0x20 DUP2 LT ISZERO PUSH2 0x10D JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST DUP2 ADD SWAP1 DUP1 DUP1 CALLDATALOAD SWAP1 PUSH1 0x20 ADD SWAP1 SWAP3 SWAP2 SWAP1 POP POP POP PUSH2 0x2F1 JUMP JUMPDEST PUSH1 0x40 MLOAD DUP1 DUP3 DUP2 MSTORE PUSH1 0x20 ADD SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 RETURN JUMPDEST CALLVALUE DUP1 ISZERO PUSH2 0x145 JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH2 0x14E PUSH2 0x415 JUMP JUMPDEST PUSH1 0x40 MLOAD DUP1 DUP3 PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 MSTORE PUSH1 0x20 ADD SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 RETURN JUMPDEST CALLVALUE DUP1 ISZERO PUSH2 0x19C JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH2 0x1A5 PUSH2 0x43B JUMP JUMPDEST PUSH1 0x40 MLOAD DUP1 DUP3 DUP2 MSTORE PUSH1 0x20 ADD SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 RETURN JUMPDEST PUSH2 0x1C3 PUSH2 0x482 JUMP JUMPDEST PUSH1 0x40 MLOAD DUP1 DUP3 DUP2 MSTORE PUSH1 0x20 ADD SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 RETURN JUMPDEST PUSH1 0x0 DUP1 ADDRESS PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND BALANCE LT ISZERO DUP1 ISZERO PUSH2 0x24F JUMPI POP PUSH1 0x2 PUSH1 0x0 SWAP1 SLOAD SWAP1 PUSH2 0x100 EXP SWAP1 DIV PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND EQ JUMPDEST ISZERO PUSH2 0x2B3 JUMPI DUP2 PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH2 0x8FC ADDRESS PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND BALANCE SWAP1 DUP2 ISZERO MUL SWAP1 PUSH1 0x40 MLOAD PUSH1 0x0 PUSH1 0x40 MLOAD DUP1 DUP4 SUB DUP2 DUP6 DUP9 DUP9 CALL SWAP4 POP POP POP POP ISZERO DUP1 ISZERO PUSH2 0x2B1 JUMPI RETURNDATASIZE PUSH1 0x0 DUP1 RETURNDATACOPY RETURNDATASIZE PUSH1 0x0 REVERT JUMPDEST POP JUMPDEST ADDRESS PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND BALANCE SWAP1 POP SWAP2 SWAP1 POP JUMP JUMPDEST PUSH1 0x0 ADDRESS PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND BALANCE SWAP1 POP SWAP1 JUMP JUMPDEST PUSH1 0x0 PUSH1 0x1 PUSH1 0x0 CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 MSTORE PUSH1 0x20 ADD SWAP1 DUP2 MSTORE PUSH1 0x20 ADD PUSH1 0x0 KECCAK256 SLOAD DUP3 GT PUSH2 0x3CE JUMPI DUP2 PUSH1 0x1 PUSH1 0x0 CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 MSTORE PUSH1 0x20 ADD SWAP1 DUP2 MSTORE PUSH1 0x20 ADD PUSH1 0x0 KECCAK256 PUSH1 0x0 DUP3 DUP3 SLOAD SUB SWAP3 POP POP DUP2 SWAP1 SSTORE POP CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH2 0x8FC DUP4 SWAP1 DUP2 ISZERO MUL SWAP1 PUSH1 0x40 MLOAD PUSH1 0x0 PUSH1 0x40 MLOAD DUP1 DUP4 SUB DUP2 DUP6 DUP9 DUP9 CALL SWAP4 POP POP POP POP ISZERO DUP1 ISZERO PUSH2 0x3CC JUMPI RETURNDATASIZE PUSH1 0x0 DUP1 RETURNDATACOPY RETURNDATASIZE PUSH1 0x0 REVERT JUMPDEST POP JUMPDEST PUSH1 0x1 PUSH1 0x0 CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 MSTORE PUSH1 0x20 ADD SWAP1 DUP2 MSTORE PUSH1 0x20 ADD PUSH1 0x0 KECCAK256 SLOAD SWAP1 POP SWAP2 SWAP1 POP JUMP JUMPDEST PUSH1 0x2 PUSH1 0x0 SWAP1 SLOAD SWAP1 PUSH2 0x100 EXP SWAP1 DIV PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 JUMP JUMPDEST PUSH1 0x0 PUSH1 0x1 PUSH1 0x0 CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 MSTORE PUSH1 0x20 ADD SWAP1 DUP2 MSTORE PUSH1 0x20 ADD PUSH1 0x0 KECCAK256 SLOAD SWAP1 POP SWAP1 JUMP JUMPDEST PUSH1 0x0 CALLVALUE PUSH1 0x1 PUSH1 0x0 CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 MSTORE PUSH1 0x20 ADD SWAP1 DUP2 MSTORE PUSH1 0x20 ADD PUSH1 0x0 KECCAK256 PUSH1 0x0 DUP3 DUP3 SLOAD ADD SWAP3 POP POP DUP2 SWAP1 SSTORE POP CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH32 0xA8126F7572BB1FDEAE5B5AA9EC126438B91F658A07873F009D041AE690F3A193 CALLVALUE PUSH1 0x40 MLOAD DUP1 DUP3 DUP2 MSTORE PUSH1 0x20 ADD SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 LOG2 PUSH1 0x1 PUSH1 0x0 CALLER PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 MSTORE PUSH1 0x20 ADD SWAP1 DUP2 MSTORE PUSH1 0x20 ADD PUSH1 0x0 KECCAK256 SLOAD SWAP1 POP SWAP1 JUMP INVALID LOG1 PUSH6 0x627A7A723058 KECCAK256 PUSH22 0x3D1CC549A04AD30E75834D6C8210A508CF241A6B16B4 0xd4 PUSH25 0xA2F206A121608C0029333020657468657220696E697469616C KECCAK256 PUSH7 0x756E64696E6720 PUSH19 0x65717569726564000000000000000000000000 ", "sourceMap": "25:1826:0:-;;;391:8;378:9;:21;370:67;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;455:10;447:5;;:18;;;;;;;;;;;;;;;;;;489:1;475:11;;:15;;;;;;;;;;;;;;;;;;25:1826;;;;;;" }`
	abiContract = `[ { "constant": false, "inputs": [ { "name": "organization", "type": "address" } ], "name": "pay", "outputs": [ { "name": "remainingBal", "type": "uint256" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "depositsBalance", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [ { "name": "withdrawAmount", "type": "uint256" } ], "name": "withdraw", "outputs": [ { "name": "remainingBal", "type": "uint256" } ], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "owner", "outputs": [ { "name": "", "type": "address" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": true, "inputs": [], "name": "balance", "outputs": [ { "name": "", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" }, { "constant": false, "inputs": [], "name": "deposit", "outputs": [ { "name": "", "type": "uint256" } ], "payable": true, "stateMutability": "payable", "type": "function" }, { "inputs": [], "payable": true, "stateMutability": "payable", "type": "constructor" }, { "anonymous": false, "inputs": [ { "indexed": true, "name": "accountAddress", "type": "address" }, { "indexed": false, "name": "amount", "type": "uint256" } ], "name": "LogDepositMade", "type": "event" } ]`
)

func TestDeployContract(t *testing.T) {
	client, err := ethclient.Dial("http://localhost:59047")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	privateKey, err := crypto.HexToECDSA("29f4455cd82770e305096f33c2a53f13efed2974873c883a6d7ca7d1bdcdf0c7")
	if err != nil {
		t.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := crypto.CreateAddress(fromAddress, nonce)

	chainID := big.NewInt(45439)

	err = SendTx(ctx, gasPrice, chainID, big.NewInt(0), &contractAddress, &fromAddress, privateKey, client, []byte(bytecode))
	if err != nil {
		t.Fatal(err)
	}

	ssAbi, _ := abi.JSON(strings.NewReader(abiContract))
	data, err := ssAbi.Pack("deposit")
	if err != nil {
		t.Fatal(err)
	}

	err = SendTx(ctx, gasPrice, chainID, big.NewInt(100), &contractAddress, &fromAddress, privateKey, client, data)
	if err != nil {
		t.Fatal(err)
	}

	ssAbi, _ = abi.JSON(strings.NewReader(abiContract))
	data, err = ssAbi.Pack("pay", common.HexToAddress("0xff492750E00A6A7Cba5e4089C542ac41E96885fB"))
	if err != nil {
		t.Fatal(err)
	}

	err = SendTx(ctx, gasPrice, chainID, big.NewInt(100), &contractAddress, &fromAddress, privateKey, client, data)
	if err != nil {
		t.Fatal(err)
	}
}

func SendTx(ctx context.Context, gasPrice, chainID, value *big.Int, toAddress, fromAddress *common.Address, prv *ecdsa.PrivateKey, client *ethclient.Client, data []byte) error {
	gas, err := client.EstimateGas(ctx, ethereum.CallMsg{From: *fromAddress, To: toAddress, GasPrice: big.NewInt(0), Data: data})
	if err != nil {
		return err
	}

	nonce, err := client.PendingNonceAt(context.Background(), *fromAddress)
	if err != nil {
		return err
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gas,
		Data:     data,
		To:       toAddress,
		Value:    value,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(chainID), prv)

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}

	txReceipt, err := WaitTickerTx(ctx, client, signedTx.Hash())
	if err != nil {
		return err
	}

	if txReceipt.Status == types.ReceiptStatusFailed {
		return errors.New("bad status")
	}

	return nil
}
