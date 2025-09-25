package scanner

import (
	"github.com/kslamph/tronlib/pb/api"
	"github.com/kslamph/tronlib/pkg/types"
	"github.com/kslamph/tronlib/pkg/utils"
)

// only use this func if the addr []byte is guaranteed to be a valid address
func byteAddrToString(addr []byte) string {
	return types.MustNewAddressFromBytes(addr).String() // validate
}

// recoverSignersFromTransaction recovers all signer addresses from transaction signatures using the tronlib utility
func recoverSignersFromTransaction(tx *api.TransactionExtention) ([]string, error) {
	if tx.Transaction == nil {
		return nil, nil
	}

	// Use the tronlib utility function to extract signers
	signerAddresses, err := utils.ExtractSigners(tx.Transaction)
	if err != nil {
		return nil, err
	}

	// Convert the types.Address objects to strings
	signers := make([]string, 0, len(signerAddresses))
	for _, addr := range signerAddresses {
		if addr != nil {
			signers = append(signers, addr.String())
		}
	}

	return signers, nil
}
