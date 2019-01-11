package blockchainlistener

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/SmartMeshFoundation/Photon/log"

	"github.com/SmartMeshFoundation/Photon-Path-Finder/model"

	"github.com/SmartMeshFoundation/Photon/utils"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
)

type channelIDStruct struct {
	p1            common.Address
	p2            common.Address
	token         common.Address
	tokensNetwork common.Address
	channelID     common.Hash
}

func TestCalcChannelID(t *testing.T) {
	model.SetupTestDB()
	cases := []channelIDStruct{
		{
			p1:            common.HexToAddress("0x4B89Bff01009928784eB7e7d10Bf773e6D166066"),
			p2:            common.HexToAddress("0x3af7fbddef2CeBEeB850328a0834Aa9a29684332"),
			token:         common.HexToAddress("0x10642C068d38f1567d97E3ED1EEAFb8c2420ff54"),
			tokensNetwork: common.HexToAddress("0x3e4D30AAba71670921C448A1951AEb0a1414ba09"),
			channelID:     common.HexToHash("0x23ac04787505ab7fd9fe0519df0b12ce4296dd6e14632f594dd195e32b20a36a"),
		},
		{
			p1:            common.HexToAddress("0x292650fee408320D888e06ed89D938294Ea42f99"),
			p2:            common.HexToAddress("0x4B89Bff01009928784eB7e7d10Bf773e6D166066"),
			token:         common.HexToAddress("0x10642C068d38f1567d97E3ED1EEAFb8c2420ff54"),
			tokensNetwork: common.HexToAddress("0x3e4D30AAba71670921C448A1951AEb0a1414ba09"),
			channelID:     common.HexToHash("0x9653fe73704182cb7b1377cfae1471a304ab94eb824979be5a22464b507dd8cc"),
		},
	}
	for _, c := range cases {
		cid := calcChannelID(c.token, c.tokensNetwork, c.p1, c.p2)
		assert.EqualValues(t, cid, c.channelID)
		cid = calcChannelID(c.token, c.tokensNetwork, c.p2, c.p1)
		assert.EqualValues(t, cid, c.channelID)
	}
}

func TestTokenNetwork_GetPaths(t *testing.T) {
	model.SetupTestDB()
	token := utils.NewRandomAddress()
	tokensNetwork := utils.NewRandomAddress()
	tn := NewTokenNetwork(nil, tokensNetwork)
	tn.decimals = map[common.Address]int{
		token: 0,
	}
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokensNetwork,
	}
	addr1, addr2, addr3 := utils.NewRandomAddress(), utils.NewRandomAddress(), utils.NewRandomAddress()
	tn.participantStatus[addr1] = nodeStatus{false, true}
	tn.participantStatus[addr2] = nodeStatus{false, true}
	tn.participantStatus[addr3] = nodeStatus{false, true}
	c1Id := calcChannelID(token, tokensNetwork, addr1, addr2)
	tn.handleChannelOpenedEvent(token, c1Id, addr1, addr2, 3)
	tn.channels[c1Id].Participant1Balance = big.NewInt(20)
	tn.channels[c1Id].Participant2Balance = big.NewInt(20)
	tn.channels[c1Id].Participant1Fee = &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: big.NewInt(1),
	}
	tn.channels[c1Id].Participant2Fee = &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: big.NewInt(1),
	}

	paths, err := tn.GetPaths(addr1, addr2, token, big.NewInt(10), 3, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 1 || paths[0].PathHop != 0 {
		t.Errorf("length should be 0,paths=%s", utils.StringInterface(paths, 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr2, token, big.NewInt(30), 3, "")
	if err == nil {
		t.Error("should no path")
		return
	}

	c2Id := calcChannelID(token, tokensNetwork, addr2, addr3)
	tn.handleChannelOpenedEvent(token, c2Id, addr2, addr3, 3)
	tn.channels[c2Id].Participant1Balance = big.NewInt(20)
	tn.channels[c2Id].Participant2Balance = big.NewInt(20)
	tn.channels[c2Id].Participant1Fee = &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: big.NewInt(1),
	}
	tn.channels[c2Id].Participant2Fee = &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: big.NewInt(1),
	}
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(3), 5, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 2 || paths[0].PathHop != 1 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(30), 5, "")
	if err == nil {
		t.Error("should not path")
		return
	}
}
func TestTokenNetwork_getWeight(t *testing.T) {
	model.SetupTestDB()
	token := utils.NewRandomAddress()
	tokenNetwork := utils.NewRandomAddress()
	tn := NewTokenNetwork(nil, tokenNetwork)
	tn.decimals = map[common.Address]int{
		token: 18,
	}
	base := big.NewInt(int64(math.Pow10(18)))
	balance := big.NewInt(20)
	balance = balance.Mul(balance, base)
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokenNetwork,
	}
	w := tn.getWeight(token, &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: big.NewInt(3000000000),
	}, big.NewInt(20))
	//because of accuracy
	assert.EqualValues(t, w, 0)
	w = tn.getWeight(token, &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: big.NewInt(300000000000000000),
	}, big.NewInt(20))
	assert.EqualValues(t, w, 30000)
	w = tn.getWeight(token, &model.Fee{
		FeePolicy:   model.FeePolicyPercent,
		FeeConstant: big.NewInt(0),
		FeePercent:  10000,
	}, big.NewInt(20))
	assert.EqualValues(t, w, 0)
	w = tn.getWeight(token, &model.Fee{
		FeePolicy:   model.FeePolicyPercent,
		FeeConstant: big.NewInt(0),
		FeePercent:  10000,
	}, big.NewInt(2000000))
	assert.EqualValues(t, w, 0)

	w = tn.getWeight(token, &model.Fee{
		FeePolicy:   model.FeePolicyPercent,
		FeeConstant: big.NewInt(0),
		FeePercent:  10000,
	}, big.NewInt(2000000000000000000))
	assert.EqualValues(t, w, 20)
}
func TestTokenNetwork_GetPathsBigInt(t *testing.T) {
	model.SetupTestDB()
	token := utils.NewRandomAddress()
	tokenNetwork := utils.NewRandomAddress()
	tn := NewTokenNetwork(nil, tokenNetwork)
	tn.decimals = map[common.Address]int{
		token: 18,
	}
	base := big.NewInt(int64(math.Pow10(18)))
	balance := big.NewInt(20)
	balance = balance.Mul(balance, base)
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokenNetwork,
	}
	addr1, addr2, addr3 := utils.NewRandomAddress(), utils.NewRandomAddress(), utils.NewRandomAddress()
	tn.participantStatus[addr1] = nodeStatus{false, true}
	tn.participantStatus[addr2] = nodeStatus{false, true}
	tn.participantStatus[addr3] = nodeStatus{false, true}
	fee := big.NewInt(1)
	fee.Mul(fee, base)

	c1Id := calcChannelID(token, tokenNetwork, addr1, addr2)
	tn.handleChannelOpenedEvent(token, c1Id, addr1, addr2, 3)
	tn.channels[c1Id].Participant1Fee = &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: fee,
	}
	tn.channels[c1Id].Participant2Fee = &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: fee,
	}
	tn.channels[c1Id].Participant1Balance = balance
	tn.channels[c1Id].Participant2Balance = balance

	v := big.NewInt(10)
	paths, err := tn.GetPaths(addr1, addr2, token, v.Mul(v, base), 3, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 1 || paths[0].PathHop != 0 {
		t.Errorf("length should be 0,paths=%s", utils.StringInterface(paths, 3))
		return
	}
	v = big.NewInt(30)
	paths, err = tn.GetPaths(addr1, addr2, token, v.Mul(v, base), 3, "")
	if err == nil {
		t.Error("should no path")
		return
	}
	c2Id := calcChannelID(token, tokenNetwork, addr2, addr3)
	tn.handleChannelOpenedEvent(token, c2Id, addr2, addr3, 3)
	tn.channels[c2Id].Participant1Fee = &model.Fee{
		FeePolicy:   model.FeePolicyCombined,
		FeeConstant: fee,
		FeePercent:  1000, //额外收千分之一
	}
	tn.channels[c2Id].Participant2Fee = &model.Fee{
		FeePolicy:   model.FeePolicyConstant,
		FeeConstant: fee,
	}
	tn.channels[c2Id].Participant1Balance = balance
	tn.channels[c2Id].Participant2Balance = balance

	v = big.NewInt(3)
	paths, err = tn.GetPaths(addr1, addr3, token, v.Mul(v, base), 5, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 2 || paths[0].PathHop != 1 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	t.Logf("paths=%s", utils.StringInterface(paths, 3))
	v = big.NewInt(30)
	paths, err = tn.GetPaths(addr1, addr3, token, v.Mul(v, base), 5, "")
	if err == nil {
		t.Error("should not path")
		return
	}
}

func TestTokenNetwork_GetPathsMultiHop(t *testing.T) {
	model.SetupTestDB()
	token := utils.NewRandomAddress()
	tokenNetwork := utils.NewRandomAddress()
	tn := NewTokenNetwork(nil, tokenNetwork)
	tn.decimals = map[common.Address]int{
		token: 0,
	}
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokenNetwork,
	}
	addr1, addr2, addr3 := utils.NewRandomAddress(), utils.NewRandomAddress(), utils.NewRandomAddress()
	addr4 := utils.NewRandomAddress()
	addr5 := utils.NewRandomAddress()
	log.Trace(fmt.Sprintf("addr1=%s,\naddr2=%s,\naddr3=%s,\naddr4=%s,\naddr5=%s", addr1.String(),
		addr2.String(), addr3.String(), addr4.String(), addr5.String()))
	tn.participantStatus[addr1] = nodeStatus{false, true}
	tn.participantStatus[addr2] = nodeStatus{false, true}
	tn.participantStatus[addr3] = nodeStatus{false, true}
	tn.participantStatus[addr4] = nodeStatus{false, true}
	tn.participantStatus[addr5] = nodeStatus{false, true}
	c1 := &channel{
		Participant1: addr1,
		Participant2: addr2,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c1Id := calcChannelID(token, tokenNetwork, addr1, addr2)
	tn.channelViews[token] = []*channel{c1}
	tn.channels[c1Id] = c1
	paths, err := tn.GetPaths(addr1, addr2, token, big.NewInt(10), 3, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 1 || paths[0].PathHop != 0 {
		t.Errorf("length should be 0,paths=%s", utils.StringInterface(paths, 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr2, token, big.NewInt(30), 3, "")
	if err == nil {
		t.Error("should no path")
		return
	}
	c2 := &channel{
		Participant1: addr2,
		Participant2: addr3,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c2Id := calcChannelID(token, tokenNetwork, addr2, addr3)
	tn.channelViews[token] = []*channel{c1, c2}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(3), 5, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 2 || paths[0].PathHop != 1 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(30), 5, "")
	if err == nil {
		t.Error("should not path")
		return
	}

	c3 := &channel{
		Participant1: addr3,
		Participant2: addr5,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c3Id := calcChannelID(token, tokenNetwork, addr3, addr5)
	tn.channelViews[token] = []*channel{c1, c2, c3}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	tn.channels[c3Id] = c3

	c4 := &channel{
		Participant1: addr4,
		Participant2: addr5,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c4Id := calcChannelID(token, tokenNetwork, addr4, addr5)
	tn.channelViews[token] = []*channel{c1, c2, c3, c4}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	tn.channels[c3Id] = c3
	tn.channels[c4Id] = c4

	c5 := &channel{
		Participant1: addr2,
		Participant2: addr4,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c5Id := calcChannelID(token, tokenNetwork, addr2, addr4)
	tn.channelViews[token] = []*channel{c1, c2, c3, c4, c5}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	tn.channels[c3Id] = c3
	tn.channels[c4Id] = c4
	tn.channels[c5Id] = c5
	//1-2-3-5 or 1-2-4-5
	paths, err = tn.GetPaths(addr1, addr5, token, big.NewInt(3), 5, "")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("paths=%s", utils.StringInterface(paths, 5))
	if len(paths[0].Result) != 3 || paths[0].PathHop != 2 || paths[0].Fee.Cmp(big.NewInt(2)) != 0 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	if len(paths) != 2 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	if len(paths[1].Result) != 3 || paths[1].PathHop != 2 || paths[1].Fee.Cmp(big.NewInt(2)) != 0 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(30), 5, "")
	if err == nil {
		t.Error("should not path")
		return
	}

}

func TestTokenNetwork_handleNewChannel(t *testing.T) {

	model.SetupTestDB()
	token := utils.NewRandomAddress()
	tokenNetwork := utils.NewRandomAddress()
	tn := NewTokenNetwork(nil, tokenNetwork)
	tn.decimals = map[common.Address]int{
		token: 0,
	}
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokenNetwork,
	}
	channid := utils.NewRandomHash()
	p1 := utils.NewRandomAddress()
	p2 := utils.NewRandomAddress()
	err := tn.handleChannelOpenedEvent(token, channid, p1, p2, 3)
	if err != nil {
		t.Error(err)
		return
	}
	c := tn.channels[channid]
	assert.EqualValues(t, c.Participant1, p1)
	assert.EqualValues(t, c.Participant2, p2)
	_, err = model.GetChannelFeeRate(channid, p1)
	if err != nil {
		t.Error(err)
		return
	}
	err = tn.handleChannelClosedEvent(channid)
	if err != nil {
		t.Error(err)
		return
	}
	err = tn.handleChannelClosedEvent(channid)
	if err == nil {
		t.Error("should error")
		return
	}
}

func BenchmarkTokenNetwork_GetPaths(b *testing.B) {
	model.SetupTestDB()
	nodesNumber := 10000
	nodes := make(map[int]common.Address)
	token := utils.NewRandomAddress()
	tokenNetwork := utils.NewRandomAddress()
	tn := NewTokenNetwork(nil, tokenNetwork)
	tn.decimals = map[common.Address]int{
		token: 18,
	}
	base := big.NewInt(int64(math.Pow10(18)))
	balance := big.NewInt(20)
	balance = balance.Mul(balance, base)
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokenNetwork,
	}
	lastAddr := utils.NewRandomAddress()
	tn.participantStatus[lastAddr] = nodeStatus{false, true}
	for i := 0; i < nodesNumber; i++ {
		nodes[i] = lastAddr
		addr := utils.NewRandomAddress()
		tn.participantStatus[addr] = nodeStatus{false, true}
		c := &channel{
			Participant1: lastAddr,
			Participant2: addr,
			Participant1Fee: &model.Fee{
				FeePercent:  model.FeePolicyConstant,
				FeeConstant: big.NewInt(1),
			},
			Participant2Fee: &model.Fee{
				FeePercent:  model.FeePolicyConstant,
				FeeConstant: big.NewInt(1),
			},
			Participant1Balance: big.NewInt(100000),
			Participant2Balance: big.NewInt(100000),
		}
		cid := calcChannelID(token, tokenNetwork, lastAddr, addr)
		cs := tn.channelViews[token]
		cs = append(cs, c)
		tn.channels[cid] = c
		tn.channelViews[token] = cs

		//for next channel
		lastAddr = addr
	}
	b.N = 100
	for i := 0; i < b.N; i++ {
		from := nodes[utils.NewRandomInt(nodesNumber)]
		to := nodes[utils.NewRandomInt(nodesNumber)]
		//go func(from, to common.Address) {
		paths, err := tn.GetPaths(from, to, token, big.NewInt(10), 5, "")
		if err != nil {
			b.Error(err)
			return
		}
		if len(paths) != 1 {
			b.Errorf("length should be 1,paths=%s", utils.StringInterface(paths, 3))
			return
		}
		//}(from, to)

	}

	/*
		1秒(s)=1000000000纳秒(ns)
			在顺序进行的情况下,占用内存2.14g(比较稳定,与N没关系,b.N=100依然如此)
															1000000000
			BenchmarkTokenNetwork_GetPaths-8   	     100	1102199349 ns/op
			在并发进行的情况下,占满这个电脑内存(16g),
			BenchmarkTokenNetwork_GetPaths-8   	     100	 306597751 ns/op
	*/
}

func TestDeleteSlice(t *testing.T) {
	var cs []int
	//cs = []int{1, 2, 3}
	for k, v := range cs {
		if v == 1 {
			cs = append(cs[:k], cs[k+1:]...)
			//break
		}
	}
}

func TestTokenNetwork_GetPaths2(t *testing.T) {
	model.SetupTestDB()
	token := utils.NewRandomAddress()
	tokenNetwork := utils.NewRandomAddress()
	tn := NewTokenNetwork(nil, tokenNetwork)
	tn.decimals = map[common.Address]int{
		token: 0,
	}
	tn.token2TokenNetwork = map[common.Address]common.Address{
		token: tokenNetwork,
	}
	addr1, addr2, addr3 := utils.NewRandomAddress(), utils.NewRandomAddress(), utils.NewRandomAddress()
	addr4 := utils.NewRandomAddress()
	addr5 := utils.NewRandomAddress()
	log.Trace(fmt.Sprintf("addr1=%s,\naddr2=%s,\naddr3=%s,\naddr4=%s,\naddr5=%s", addr1.String(),
		addr2.String(), addr3.String(), addr4.String(), addr5.String()))
	tn.participantStatus[addr1] = nodeStatus{false, true}
	tn.participantStatus[addr2] = nodeStatus{false, true}
	tn.participantStatus[addr3] = nodeStatus{false, true}
	tn.participantStatus[addr4] = nodeStatus{false, true}
	tn.participantStatus[addr5] = nodeStatus{false, true}
	c1 := &channel{
		Participant1: addr1,
		Participant2: addr2,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c1Id := calcChannelID(token, tokenNetwork, addr1, addr2)
	tn.channelViews[token] = []*channel{c1}
	tn.channels[c1Id] = c1
	paths, err := tn.GetPaths(addr1, addr2, token, big.NewInt(10), 3, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 1 || paths[0].PathHop != 0 {
		t.Errorf("length should be 0,paths=%s", utils.StringInterface(paths, 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr2, token, big.NewInt(30), 3, "")
	if err == nil {
		t.Error("should no path")
		return
	}
	c2 := &channel{
		Participant1: addr2,
		Participant2: addr3,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c2Id := calcChannelID(token, tokenNetwork, addr2, addr3)
	tn.channelViews[token] = []*channel{c1, c2}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(3), 5, "")
	if err != nil {
		t.Error(err)
		return
	}
	if len(paths[0].Result) != 2 || paths[0].PathHop != 1 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	paths, err = tn.GetPaths(addr1, addr3, token, big.NewInt(30), 5, "")
	if err == nil {
		t.Error("should not path")
		return
	}

	c3 := &channel{
		Participant1: addr3,
		Participant2: addr5,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c3Id := calcChannelID(token, tokenNetwork, addr3, addr5)
	tn.channelViews[token] = []*channel{c1, c2, c3}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	tn.channels[c3Id] = c3

	//c4 := &channel{
	//	Participant1: addr4,
	//	Participant2: addr5,
	//	Participant1Fee: &model.Fee{
	//		FeePolicy:   model.FeePolicyConstant,
	//		FeeConstant: big.NewInt(1),
	//	},
	//	Participant2Fee: &model.Fee{
	//		FeePolicy:   model.FeePolicyConstant,
	//		FeeConstant: big.NewInt(1),
	//	},
	//	Participant1Balance: big.NewInt(20),
	//	Participant2Balance: big.NewInt(20),
	//}
	//c4Id := calcChannelID(token, tokenNetwork, addr4, addr5)
	//tn.channelViews[token] = []*channel{c1, c2, c3, c4}
	//tn.channels[c2Id] = c2
	//tn.channels[c1Id] = c1
	//tn.channels[c3Id] = c3
	//tn.channels[c4Id] = c4

	c5 := &channel{
		Participant1: addr2,
		Participant2: addr5,
		Participant1Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(10),
		},
		Participant2Fee: &model.Fee{
			FeePolicy:   model.FeePolicyConstant,
			FeeConstant: big.NewInt(1),
		},
		Participant1Balance: big.NewInt(20),
		Participant2Balance: big.NewInt(20),
	}
	c5Id := calcChannelID(token, tokenNetwork, addr2, addr5)
	tn.channelViews[token] = []*channel{c1, c2, c3, c5}
	tn.channels[c2Id] = c2
	tn.channels[c1Id] = c1
	tn.channels[c3Id] = c3
	//tn.channels[c4Id] = c4
	tn.channels[c5Id] = c5
	//1-2-3-5 or 1-2-4-5
	paths, err = tn.GetPaths(addr1, addr5, token, big.NewInt(3), 5, "")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("paths=%s", utils.StringInterface(paths, 5))
	if len(paths[0].Result) != 3 || paths[0].PathHop != 2 || paths[0].Fee.Cmp(big.NewInt(2)) != 0 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}
	if len(paths) != 1 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}

	paths, err = tn.GetPaths(addr2, addr5, token, big.NewInt(3), 5, "")
	if err != nil {
		t.Error("should have one path")
		return
	}
	if len(paths[0].Result) != 1 || paths[0].PathHop != 0 || paths[0].Fee.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("path length error,paths=%s", utils.StringInterface(paths[0], 3))
		return
	}

}
