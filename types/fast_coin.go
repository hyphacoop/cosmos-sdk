package types

import (
	"errors"
	"fmt"
	"math/big"
	"sync"

	"cosmossdk.io/math"
)

// FastCoin is a high-performance alternative to sdk.Coin that uses object pooling
// to reduce memory allocations in hot paths
type FastCoin struct {
	Denom     string
	Amount    math.Int
	pool      *FastCoinPool
	intPooled bool
}

// FastCoinPool manages a pool of FastCoin objects
type FastCoinPool struct {
	pool    sync.Pool
	intPool sync.Pool
}

// NewFastCoinPool creates a new FastCoinPool
func NewFastCoinPool() *FastCoinPool {
	return &FastCoinPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &FastCoin{}
			},
		},
		intPool: sync.Pool{
			New: func() interface{} {
				return new(big.Int)
			},
		},
	}
}

// Get returns a FastCoin from the pool
func (p *FastCoinPool) Get(denom string, amount math.Int) *FastCoin {
	coin := p.pool.Get().(*FastCoin)
	coin.Denom = denom
	coin.Amount = amount
	coin.pool = p
	coin.intPooled = false
	return coin
}

func (f *FastCoin) Release() {
	if f.intPooled {
		f.pool.intPool.Put(f.Amount.BigIntMut())
		f.intPooled = false
	}
	f.pool.pool.Put(f)
}

func (f *FastCoin) subAmount(amount *math.Int) math.Int {
	newInt := f.pool.intPool.Get().(*big.Int)
	newInt.Sub(f.Amount.BigIntMut(), amount.BigIntMut())
	return math.NewIntFromBigIntMut(newInt)
}

func (f *FastCoin) addAmount(amount *math.Int) math.Int {
	newInt := f.pool.intPool.Get().(*big.Int)
	newInt.Add(f.Amount.BigIntMut(), amount.BigIntMut())
	return math.NewIntFromBigIntMut(newInt)
}

// SafeSub performs subtraction and returns whether result would be negative
func (f *FastCoin) SafeSub(other *FastCoin) (result *FastCoin, hasNeg bool) {
	if f.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", f.Denom, other.Denom))
	}

	newAmount := f.subAmount(&other.Amount)

	// Check if result would be negative
	if newAmount.IsNegative() {
		f.pool.intPool.Put(newAmount.BigIntMut())
		return nil, true
	}

	// Get a new coin from the pool for the result
	result = f.pool.Get(f.Denom, newAmount)
	result.intPooled = true
	return result, false
}

func (f *FastCoin) Sub(other *FastCoin) *FastCoin {
	result, hasNeg := f.SafeSub(other)
	if hasNeg {
		panic("negative coin amount")
	}
	return result
}

func (f *FastCoin) SafeSubCoin(other *Coin) (*FastCoin, bool) {
	if f.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", f.Denom, other.Denom))
	}
	newAmount := f.subAmount(&other.Amount)
	if newAmount.IsNegative() {
		f.pool.intPool.Put(newAmount.BigIntMut())
		return nil, true
	}
	result := f.pool.Get(f.Denom, newAmount)
	return result, false
}

func (f *FastCoin) SubCoin(other *Coin) *FastCoin {
	result, hasNeg := f.SafeSubCoin(other)
	if hasNeg {
		panic("negative coin amount")
	}
	return result
}

func (f *FastCoin) String() string {
	return fmt.Sprintf("%s%s", f.Amount, f.Denom)
}

func (f *FastCoin) IsZero() bool {
	return f.Amount.IsZero()
}

func (f *FastCoin) Validate() error {
	if err := ValidateDenom(f.Denom); err != nil {
		return err
	}
	if f.Amount.IsNil() {
		return errors.New("coin amount is nil")
	}
	if f.Amount.IsNegative() {
		return errors.New("coin amount is negative")
	}
	return nil
}

func (f *FastCoin) IsValid() bool {
	return f.Validate() == nil
}

func (f *FastCoin) Add(other *FastCoin) *FastCoin {
	if f.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", f.Denom, other.Denom))
	}
	newAmount := f.addAmount(&other.Amount)
	newCoin := f.pool.Get(f.Denom, newAmount)
	newCoin.intPooled = true
	return newCoin
}

func (f *FastCoin) AddCoin(other *Coin) *FastCoin {
	if f.Denom != other.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", f.Denom, other.Denom))
	}
	newAmount := f.addAmount(&other.Amount)
	newCoin := f.pool.Get(f.Denom, newAmount)
	newCoin.intPooled = true
	return newCoin
}
