package core

import "time"

// IndexComponent defines a single asset within an index token.
type IndexComponent struct {
	AssetID  TokenID
	Weight   float64
	Quantity uint64
}

// IndexToken extends BaseToken with index specific metadata.
type IndexToken struct {
	*BaseToken
	components    []IndexComponent
	lastRebalance time.Time
	marketValue   uint64
}

// NewIndexToken creates a new SYN3700 index token with initial components.
func NewIndexToken(meta Metadata, comps []IndexComponent, ledger *Ledger, gas GasCalculator) (*IndexToken, error) {
	tok, err := (Factory{}).Create(meta, map[Address]uint64{})
	if err != nil {
		return nil, err
	}
	bt := tok.(*BaseToken)
	it := &IndexToken{BaseToken: bt, components: comps, lastRebalance: meta.Created}
	bt.ledger = ledger
	bt.gas = gas
	if ledger.tokens == nil {
		ledger.tokens = make(map[TokenID]Token)
	}
	ledger.tokens[bt.id] = it
	RegisterToken(it)
	return it, nil
}

// Rebalance replaces the index composition and updates the timestamp.
func (it *IndexToken) Rebalance(comps []IndexComponent) {
	it.components = comps
	it.lastRebalance = time.Now().UTC()
}

// UpdateMarketValue sets the current market value for the index.
func (it *IndexToken) UpdateMarketValue(v uint64) { it.marketValue = v }

// Components returns the current list of index components.
func (it *IndexToken) Components() []IndexComponent { return it.components }

// MarketValue returns the last recorded market value.
func (it *IndexToken) MarketValue() uint64 { return it.marketValue }

// LastRebalance returns the timestamp of the last rebalance operation.
func (it *IndexToken) LastRebalance() time.Time { return it.lastRebalance }
