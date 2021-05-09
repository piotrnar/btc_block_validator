package main

import (
	"time"
	"os"
	"net/http"
	"io/ioutil"
	"strings"
	"github.com/piotrnar/gocoin/lib/btc"
	"github.com/piotrnar/gocoin/lib/chain"
	"math/big"
)

const DatabaseDir = "chain/"

var BlockChain *chain.Chain

type my_handler struct {
}

func rpc_result(e error) string {
	s := e.Error()
	i := strings.Index(s, "RPC_Result:")
	if i!=-1 {
		return s[i+11:]
	}
	return ""
}


var cache = make(map[[8]byte] *btc.Block)

func cache_block(bl *btc.Block) {
	cache[bl.Hash.BIdx()] = bl
}

func redo_cached_blocks() bool {
	for k, block := range cache {
		er, _, maybelater := BlockChain.CheckBlock(block)
		if er != nil {
			if !maybelater {
				delete(cache, k)
				return true
			}
			continue
		}
		er = BlockChain.AcceptBlock(block)
		if er == nil {
			println("accepted block from cache", block.Hash.String())
		} else {
			println("rejected block from cache", block.Hash.String())
		}
		delete(cache, k)
		return true
	}
	return false
}


func (h my_handler) ServeHTTP(wr http.ResponseWriter, re *http.Request) {
	bl, _ := ioutil.ReadAll(re.Body)

	block, er := btc.NewBlock(bl)
	if er != nil {
		wr.Write([]byte(er.Error()))
		return
	}

	must_connect := re.FormValue("connect")=="true"
	expected_top_after := btc.NewUint256FromString(re.FormValue("newtop"))
	bid := strings.Trim(re.FormValue("blockid"), "\"")

	defer func() {
		last := BlockChain.LastBlock()
		println("last_block:", last.BlockHash.String(), last.Height, len(cache))
		if last.BlockHash.Equal(expected_top_after) {
			wr.Write([]byte("ok"))
			if bid=="b1004" {
				println("All tests PASSED")
				go func() { // start it in a gouroutine so http server would still send the response
					time.Sleep(1e9)
					BlockChain.Close()
					os.RemoveAll(DatabaseDir)
					os.Exit(0)
				}()
			}
		} else {
			println("expected_block:", expected_top_after.String(), re.FormValue("newheight"))
			wr.Write([]byte("error"))
		}
	}()

	println()
	println("NewBlock", bid)

	if bid=="b61" {
		println("Ignore test b61 as it won't happen in the real world anymore")
		return
	}

	//println("Expected result:", re.FormValue("connect"))
	//println("Expected exception:", re.FormValue("exception"))
	//println("Expected new top:", re.FormValue("newtop"))
	//println("Expected new height:", re.FormValue("newheight"))
	//println("Data length:", len(bl))

	er, _, maybelater := BlockChain.CheckBlock(block)

	if er != nil {
		rpc_res := rpc_result(er)
		//println(er.Error())
		if rpc_res=="" {
			wr.Write([]byte(er.Error()))
			return
		}
		if rpc_res=="duplicate" {
			return
		}
		if !must_connect {
			if maybelater {
				println("maybe later")
				cache_block(block)
			}
			return
		}
		return
	}

	er = BlockChain.AcceptBlock(block)

	if er == nil {
		println("Block accepted. Redo cached blocks", len(cache), "...")
		for redo_cached_blocks() {
		}
	}

	return
}


func main() {
	os.RemoveAll(DatabaseDir)
	BlockChain = chain.NewChainExt(DatabaseDir,
		btc.NewUint256FromString("0f9188f13cb7b2c71f2a335e3a4fc328bf5beb436012afca590b1a11466e2206"),
		true, &chain.NewChanOpts{}, nil/*&chain.BlockDBOpts{MaxCachedBlocks:500}*/)

	BlockChain.Unspent.UnwindBufLen = 1200

	BlockChain.Consensus.MaxPOWBits = 0x207fffff
	BlockChain.Consensus.GensisTimestamp = 1296688602
	BlockChain.Consensus.MaxPOWValue, _ = new(big.Int).SetString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	BlockChain.RebuildGenesisHeader() // Update Timestamp in the genesis header

	http.ListenAndServe("127.0.0.1:18444", new(my_handler))

	BlockChain.Close()
	os.RemoveAll(DatabaseDir)
}
