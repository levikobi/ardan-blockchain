package blockchain

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"
)

// zeroHash represents a has code of zeros.
const zeroHash string = "00000000000000000000000000000000"

// =============================================================================

// BlockHeader represents common information required for each block.
type BlockHeader struct {
	ParentHash  string `json:"parent_hash"` // The hash of the previous block in the chain.
	Beneficiary string `json:"beneficiary"` // The account who receives the reward and gas fee.
	Difficulty  int    `json:"difficulty"`  // The number of 0's needed to solve the hash solution.
	Number      uint64 `json:"number"`      // The block number in the chain.
	GasPrice    uint   `json:"gas_price"`   // The actual amount of gas spent to execute the transaction.
	GasLimit    uint   `json:"gas_limit"`   // The max amount of gas associated with the transaction.
	TimeStamp   uint64 `json:"timestamp"`   // The time the block was mined.
	Nonce       uint64 `json:"nonce"`       // The value identified to solve the hash solution.
}

// Block represents a set of transactions grouped together.
type Block struct {
	Header       BlockHeader `json:"header"`
	Transactions []Tx        `json:"txs"`
}

// NewBlock constructs a new BlockFS for persisting.
func NewBlock(beneficiary string, difficulty int, parentBlock Block, txs map[ID]Tx) Block {
	parentHash := zeroHash
	if parentBlock.Header.Number > 0 {
		parentHash = parentBlock.Hash()
	}

	// Store a copy of the transactions so the mempool is not
	// affected until the block is mined.
	cpy := make([]Tx, 0, len(txs))
	for _, tx := range txs {
		cpy = append(cpy, tx)
	}

	return Block{
		Header: BlockHeader{
			ParentHash:  parentHash,
			Beneficiary: beneficiary,
			Difficulty:  difficulty,
			Number:      parentBlock.Header.Number + 1,
			TimeStamp:   uint64(time.Now().Unix()),
		},
		Transactions: cpy,
	}
}

// Hash returns the unique hash for the block by marshaling
// the block into JSON and performing a hashing operation.
func (b Block) Hash() string {
	if b.Header.Number == 0 {
		return zeroHash
	}

	data, err := json.Marshal(b)
	if err != nil {
		return zeroHash
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// =============================================================================

// BlockFS represents what is written to the DB file.
type BlockFS struct {
	Hash  string
	Block Block
}

// performPOW does the work of mining to find a valid hash for a specified
// block and returns a BlockFS ready to be written to disk.
func performPOW(ctx context.Context, difficulty int, b Block, ev EventHandler) (BlockFS, time.Duration, error) {
	t := time.Now()

	// Choose a random starting point for the nonce.
	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return BlockFS{}, time.Since(t), ctx.Err()
	}
	b.Header.Nonce = nBig.Uint64()

	var attempts uint64
	for {
		attempts++
		if attempts%1_000_000 == 0 {
			ev("bcWorker: runMiningOperation: **********: miningG: mining attempts[%d]", attempts)
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			return BlockFS{}, time.Since(t), ctx.Err()
		}

		// Hash the block and check if we have solved the puzzle.
		hash := b.Hash()
		if !isHashSolved(difficulty, hash) {

			// I may want to track these nonce's to make sure I
			// don't try the same one twice.
			b.Header.Nonce++
			continue
		}

		// Did we timeout trying to solve the problem.
		if ctx.Err() != nil {
			return BlockFS{}, time.Since(t), ctx.Err()
		}

		ev("bcWorker: runMiningOperation: **********: miningG: mining final attempts[%d]", attempts)

		// We found a solution to the POW.
		bfs := BlockFS{
			Hash:  hash,
			Block: b,
		}
		return bfs, time.Since(t), nil
	}
}

// isHashSolved checks the hash to make sure it complies with
// the POW rules. We need to match a difficulty number of 0's.
func isHashSolved(difficulty int, hash string) bool {
	const match = "00000000000000000"

	if len(hash) != 64 {
		return false
	}

	return hash[:difficulty] == match[:difficulty]
}

// =============================================================================

// loadBlocksFromDisk the current set of blocks/transactions. In a real
// world situation this would require a lot of memory.
func loadBlocksFromDisk(dbPath string) ([]Block, error) {
	dbFile, err := os.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer dbFile.Close()

	var blockNum int
	var blocks []Block
	scanner := bufio.NewScanner(dbFile)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blockFS BlockFS
		if err := json.Unmarshal(scanner.Bytes(), &blockFS); err != nil {
			return nil, err
		}

		if blockFS.Block.Hash() != blockFS.Hash {
			return nil, fmt.Errorf("block %d has been changed", blockNum)
		}

		blocks = append(blocks, blockFS.Block)
		blockNum++
	}

	return blocks, nil
}