// Simple SISO Channel interface that creates links and emulates multipath channel between transmitters and receivers
// Will soon be moved to github.com/wiless/gocomm package
package channel

import (
	cell "github.com/wiless/cellular"
	"github.com/wiless/gocomm/core"
	"github.com/wiless/vlib"
	"log"
)

func init() {
	log.Println("Initiated cellular.channel")
}

type SFN struct {
	links    []cell.LinkMetric
	chparams [][]core.ChannelParam
	freqGHz  float64
}

func (s *SFN) CreateDefaultPDP() {
	s.chparams = make([][]core.ChannelParam, len(s.links))
	for i := 0; i < len(s.chparams); i++ {
		s.chparams[i] = make([]core.ChannelParam, len(s.links[i].TxNodeIDs))
		for j, _ := range s.links[i].TxNodeIDs {
			s.chparams[i][j] = core.DefaultChannel()
		}
		log.Printf("\n%d @ %f :  %#v", s.links[i].RxNodeID, s.links[i].FreqInGHz, s.chparams[i])
	}
}

type Channel struct {
	sflinks     []SFN
	freqs       vlib.VectorF
	freqindxMap map[float64]vlib.VectorI
}

func (c *Channel) CreateFromFile(file string) {
	var tmplinks []cell.LinkMetric

	c.freqindxMap = make(map[float64]vlib.VectorI)
	vlib.LoadStructure(file, &tmplinks)
	for i, v := range tmplinks {
		index := c.freqindxMap[v.FreqInGHz]
		index.AppendAtEnd(i)
		c.freqindxMap[v.FreqInGHz] = index
	}

	// log.Println(c.freqindxMap)
	c.classifySFN(tmplinks)

}

/// to be called only when freqindxMap is created
func (c *Channel) classifySFN(links []cell.LinkMetric) {

	c.sflinks = make([]SFN, len(c.freqindxMap))
	c.freqs = vlib.NewVectorF(len(c.freqindxMap))
	var i int = 0
	for f, ivec := range c.freqindxMap {
		c.sflinks[i].links = make([]cell.LinkMetric, len(ivec))
		c.sflinks[i].freqGHz = f
		c.freqs[i] = f
		for j, v := range ivec {
			c.sflinks[i].links[j] = links[v]
		}
		c.sflinks[i].CreateDefaultPDP()
		// log.Println("=================== ", f)
		// log.Println(c.sflinks[i])
		i++
	}

}

func (c *Channel) NofSFN() int {
	return c.freqs.Size()
}

/// After loading all links this must be last func to be called before running the channel
func (c *Channel) Init() {

}
