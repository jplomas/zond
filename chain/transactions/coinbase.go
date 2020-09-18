package transactions

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theQRL/zond/config"
	"github.com/theQRL/zond/crypto/dilithium"
	"github.com/theQRL/zond/protos"
	"github.com/theQRL/zond/state"

	"github.com/theQRL/qrllib/goqrllib/goqrllib"
	"github.com/theQRL/zond/misc"
)

type CoinBase struct {
	ProtocolTransaction
}

func (tx *CoinBase) BlockProposerReward() uint64 {
	return tx.pbData.GetCoinBase().GetBlockProposerReward()
}

func (tx *CoinBase) AttestorReward() uint64 {
	return tx.pbData.GetCoinBase().GetAttestorReward()
}

func (tx *CoinBase) FeeReward() uint64 {
	return tx.pbData.GetCoinBase().FeeReward
}

func (tx *CoinBase) TotalAmounts(numberOfAttestors uint64) uint64 {
	totalAmount := tx.BlockProposerReward() + tx.AttestorReward() * numberOfAttestors
	return totalAmount
}

func (tx *CoinBase) GetSigningHash(blockSigningHash []byte) []byte {
	tmp := new(bytes.Buffer)
	binary.Write(tmp, binary.BigEndian, blockSigningHash)
	binary.Write(tmp, binary.BigEndian, tx.NetworkID())
	binary.Write(tmp, binary.BigEndian, tx.Nonce())
	// PK is considered as Block proposer
	// and attestor need to sign the unsigned coinbase transaction with
	binary.Write(tmp, binary.BigEndian, tx.PK())

	binary.Write(tmp, binary.BigEndian, tx.BlockProposerReward())
	binary.Write(tmp, binary.BigEndian, tx.AttestorReward())

	tmpTXHash := misc.NewUCharVector()
	tmpTXHash.AddBytes(tmp.Bytes())
	tmpTXHash.New(goqrllib.Sha2_256(tmpTXHash.GetData()))

	return tmpTXHash.GetBytes()
}

func (tx *CoinBase) GetUnsignedHash() []byte {
	tmp := new(bytes.Buffer)
	binary.Write(tmp, binary.BigEndian, tx.NetworkID())
	binary.Write(tmp, binary.BigEndian, tx.Nonce())
	// PK is considered as Block proposer
	// and attestor need to sign the unsigned coinbase transaction with
	binary.Write(tmp, binary.BigEndian, tx.PK())

	binary.Write(tmp, binary.BigEndian, tx.BlockProposerReward())
	binary.Write(tmp, binary.BigEndian, tx.AttestorReward())

	tmpTXHash := misc.NewUCharVector()
	tmpTXHash.AddBytes(tmp.Bytes())
	tmpTXHash.New(goqrllib.Sha2_256(tmpTXHash.GetData()))

	return tmpTXHash.GetBytes()
}

func (tx *CoinBase) validateData(stateContext *state.StateContext) bool {
	txHash := tx.TxHash(tx.GetSigningHash(stateContext.BlockSigningHash()))

	coinBaseAddress := config.GetDevConfig().Genesis.CoinBaseAddress
	addressState, err := stateContext.GetAddressState(misc.Bin2HStr(coinBaseAddress))
	if err != nil {
		log.Warn("Transfer [%s] Address %s missing into state context", coinBaseAddress)
		return false
	}

	if tx.Nonce() != addressState.Nonce() + 1 {
		log.Warn(fmt.Sprintf("CoinBase [%s] Invalid Nonce %d, Expected Nonce %d",
			misc.Bin2HStr(txHash), tx.Nonce(), addressState.Nonce() + 1))
		return false
	}

	// TODO: provide total number of attestors for this check
	//balance := addressState.Balance()
	//if balance < tx.TotalAmounts() {
	//	log.Warn("Insufficient balance",
	//		"txhash", misc.Bin2HStr(txHash),
	//		"balance", balance,
	//		"fee", tx.FeeReward())
	//	return false
	//}

	//ds := stateContext.GetDilithiumState(misc.Bin2HStr(tx.PK()))
	//if ds == nil {
	//	log.Warn("Dilithium State not found for %s", misc.Bin2HStr(tx.PK()))
	//	return false
	//}
	//if !ds.Stake() {
	//	log.Warn("Dilithium PK %s is not allowed to stake", misc.Bin2HStr(tx.PK()))
	//	return false
	//}

	// TODO: Check the block proposer and attestor reward
	return true
}

func (tx *CoinBase) Validate(stateContext *state.StateContext) bool {
	signedMessage := tx.GetSigningHash(stateContext.BlockSigningHash())
	txHash := tx.TxHash(signedMessage)

	if !dilithium.DilithiumVerify(tx.Signature(), tx.PK(), signedMessage) {
		log.Warn(fmt.Sprintf("Dilithium Signature Verification failed for CoinBase Txn %s",
			misc.Bin2HStr(txHash)))
		return false
	}

	if !tx.validateData(stateContext) {
		log.Warn("Data validation failed")
		return false
	}

	return true
}

func (tx *CoinBase) ApplyStateChanges(stateContext *state.StateContext) error {
	/*
	CoinBase signature will be a dilithium signature, so it is required
	to separate this transaction into a separate coinbase transaction
	TODO:
	1. Verify signature from Dilithium Address
	 */
	//strAddrTo := misc.Bin2Qaddress(tx.AddrTo())
	//if addrState, ok := addressesState[strAddrTo]; ok {
	//	addrState.AddBalance(tx.Amount())
	//	if tx.config.Dev.RecordTransactionHashes {
			//Disabled Tracking of Transaction Hash into AddressState
			//addrState.AppendTransactionHash(tx.Txhash())
		//}
	//}

	//strAddrFrom := misc.Bin2Qaddress(tx.config.Dev.Genesis.CoinbaseAddress)
	//
	//if addrState, ok := addressesState[strAddrFrom]; ok {
	//	masterQAddr := misc.Bin2Qaddress(tx.MasterAddr())
	//	addressesState[masterQAddr].SubtractBalance(tx.Amount())
	//	if tx.config.Dev.RecordTransactionHashes {
			//Disabled Tracking of Transaction Hash into AddressState
			//addressesState[masterQAddr].AppendTransactionHash(tx.Txhash())
		//}
		//addrState.IncreaseNonce()
	//}

	// TODO:
	// Remove Block proposer and attestor address from CoinBase Txn
	// StateContext must have block proposer and attestors list
	// Coinbase must access those data from stateContext and add reward

	//txHash := tx.TxHash(tx.GetSigningHash(stateContext.BlockSigningHash()))

	//addressState, err := stateContext.GetAddressState(misc.Bin2HStr(stateContext.BlockProposer()))
	//if err != nil {
	//	return err
	//}

	//addressState.AddBalance(tx.BlockProposerReward())

	validatorsToXMSSAddress := stateContext.ValidatorsToXMSSAddress()

	strBlockProposerDilithiumPK := misc.Bin2HStr(stateContext.BlockProposer())
	// TODO: Get list of attestors
	for validatorDilithiumPK, xmssAddress  := range validatorsToXMSSAddress {
		addressState, err := stateContext.GetAddressState(misc.Bin2HStr(xmssAddress))
		if err != nil {
			return err
		}

		if validatorDilithiumPK == strBlockProposerDilithiumPK {
			addressState.AddBalance(tx.BlockProposerReward())
		} else {
			addressState.AddBalance(tx.AttestorReward())
		}
	}

	addressState, err := stateContext.GetAddressState(misc.Bin2HStr(config.GetDevConfig().Genesis.CoinBaseAddress))
	if err != nil {
		return err
	}
	addressState.SubtractBalance(tx.TotalAmounts(uint64(len(validatorsToXMSSAddress))))
	return nil
}

func (tx *CoinBase) SetAffectedAddress(stateContext *state.StateContext) error {
	coinBaseAddress := config.GetDevConfig().Genesis.CoinBaseAddress
	err := stateContext.PrepareAddressState(misc.Bin2HStr(coinBaseAddress))
	if err != nil {
		return err
	}

	// TODO: PK is dilithium PK and it must be checked if its allowed to stake current block
	//err = stateContext.PrepareAddressState(misc.Bin2HStr(misc.PK2BinAddress(tx.PK())))
	//if err != nil {
	//	return err
	//}

	err = stateContext.PrepareAddressState(misc.Bin2HStr(tx.PK()))
	if err != nil {
		return err
	}

	//for _, attestor := range tx.Attestors() {
	//	err = stateContext.PrepareAddressState(misc.Bin2HStr(attestor))
	//	if err != nil {
	//		return err
	//	}
	//}

	return err
}

func NewCoinBase(networkId uint64, blockProposer []byte, blockProposerReward uint64,
	attestorReward uint64, feeReward uint64, lastCoinBaseNonce uint64) *CoinBase {
	tx := &CoinBase{}

	tx.pbData = &protos.ProtocolTransaction{}
	tx.pbData.Type = &protos.ProtocolTransaction_CoinBase{CoinBase: &protos.CoinBase{}}

	// TODO: Derive Network ID based on the current connected network
	tx.pbData.NetworkId = networkId
	tx.pbData.Pk = blockProposer
	//tx.pbData.MasterAddr = tx.config.Dev.Genesis.CoinBaseAddress

	cb := tx.pbData.GetCoinBase()
	cb.BlockProposerReward = blockProposerReward
	cb.AttestorReward = attestorReward
	cb.FeeReward = feeReward

	// TODO: Make nonce for CoinBase sequential, as this will not be sequential due to slotNumber
	tx.pbData.Nonce = lastCoinBaseNonce + 1

	// TODO: Pass StateContext
	//if !tx.Validate(nil) {
	//	return nil
	//}

	return tx
}