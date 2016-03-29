package quantile

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var TestArray = [...]int64{9193603307515863, 6754367334490737, 3319615827007736, 5889610372829623, 9056711574806808, 552001800400511, 13400109297856412, 9290419307253104, 22335189735651727, 20198755397426237, 6750759924055697, 5282418092338629, 7669954555439878, 9871073458267054, 9596813405908124, 17826985234078452, 20809685535785593, 169504538801585, 3742572289274816, 5269690153609658, 5911223635170287, 8307883610131084, 2782048181817819, 3730609951306336, 4049109332077119, 25526756656192197, 9178098692976052, 14247068431339682, 9790147915267598, 26879545051663945, 6061096718141231, 18513304861202883, 36038916238811566, 25585683826277574, 9896041109367753, 6071196089481084, 7192619690997925, 4638251828961559, 3042642059231756, 4695726824572235, 44887308998594665, 3933519497422833, 17471932249059073, 4220651987288255, 6037514377711865, 9334562578999125, 1564928248590489, 3263044748940436, 4590412317155934, 5470186478741421, 7945869119417566, 5954581154568589, 894968515670638, 7620414993900439, 9670709128536947, 6566981438432174, 7097246010376681, 38073653883035684, 28936241604793246, 9572137207232106, 8763202610627523, 24879815234751902, 29520475378617226, 4325698909679998, 1385395600985416, 21587377639944705, 23838639891325397, 47092556792851724, 380708162070026, 7041179813266126, 11313259703460376, 16338073267858633, 7157789354354156, 7679321360662391, 1296853853173733, 8289135153494715, 5067591593829185, 6875574194262736, 6816612766339816, 2744231634144031, 3371312301999041, 3582026995577731, 9560887595174091, 9478826709460911, 14554548042944748, 2381196435702264, 920004969814432, 16786031463171992, 9843289339356544, 4813972590910607, 8148484660815154, 3645114654798244, 57439089131166135, 8901361618305092, 1776157835648735, 5659433639754501, 43703315824233925, 760807861832865, 4575552020357635, 29239935027624495}

/*
FIXME? Right now we're implementing the "lower" interpolation which is not the "linear" default one.

In [7]: np.percentile(TestArray, 50, interpolation='lower')
Out[7]: 7192619690997925

In [8]: np.percentile(TestArray, 95, interpolation='lower')
Out[8]: 36038916238811566

In [9]: np.percentile(TestArray, 99, interpolation='lower')
Out[9]: 47092556792851724

In [10]: np.percentile(TestArray, 99.9, interpolation='lower')
Out[10]: 47092556792851724
*/

// NewSummaryWithTestData returns the Summary
func NewSummaryWithTestData() *Summary {
	s := NewSummary()

	for i, v := range TestArray {
		s.Insert(v, uint64(i))
	}

	return s
}

func TestSummaryInsertion(t *testing.T) {
	assert := assert.New(t)

	s := NewSummaryWithTestData()
	assert.Equal(100, s.N)
}

type Quantile struct {
	Value   int64
	Samples []uint64
}

func TestSummaryQuantile(t *testing.T) {
	assert := assert.New(t)

	// For 100 elts and eps=0.01 the error on the rank is 0.01 * 100 = 1
	// So we can have these results:
	// *  7157789354354156, (SID=72)
	// *  7192619690997925, (SID=36)
	// *  7620414993900439, (SID=53)
	s := NewSummaryWithTestData()

	v, samples := s.Quantile(0.5)
	// our sample array only yields a sample per value
	assert.Equal(1, len(samples))
	acceptable := []Quantile{
		Quantile{Value: 7157789354354156, Samples: []uint64{72}},
		Quantile{Value: 7192619690997925, Samples: []uint64{36}},
		Quantile{Value: 7620414993900439, Samples: []uint64{53}},
	}
	foundCorrectQuantile := false
	for _, q := range acceptable {
		foundCorrectQuantile = q.Value == v && q.Samples[0] == samples[0]
		if foundCorrectQuantile {
			break
		}
	}

	assert.True(foundCorrectQuantile, "Quantile %d (samples=%v) not found", v, samples)
}

func BenchmarkSummaryInsertion(b *testing.B) {
	s := NewSummary()
	for n := 0; n < b.N; n++ {
		val := rand.Int63()
		s.Insert(val, uint64(n))
	}
}

func TestSummaryMarshal(t *testing.T) {
	assert := assert.New(t)

	s := NewSummaryWithTestData()

	b, err := json.Marshal(s)
	assert.Nil(err)

	// Now test contents
	ss := Summary{}
	err = json.Unmarshal(b, &ss)

	assert.Equal(s.N, ss.N)
	v1, samp1 := s.Quantile(0.5)
	v2, samp2 := ss.Quantile(0.5)

	assert.Equal(v1, v2)
	assert.Equal(1, len(samp1))
	assert.Equal(1, len(samp2))

	// Verify samples are correct
	samp1Correct := false
	for i, val := range TestArray {
		if val == v1 && samp1[0] == uint64(i) {
			samp1Correct = true
			break
		}
	}
	assert.True(samp1Correct, "1: sample %v incorrect for quantile %d", samp1, v1)

	samp2Correct := false
	for i, val := range TestArray {
		if val == v2 && samp2[0] == uint64(i) {
			samp2Correct = true
			break
		}
	}
	assert.True(samp2Correct, "2: sample %v incorrect for quantile %d", samp2, v2)
}

func TestSummaryMerge(t *testing.T) {
	assert := assert.New(t)

	s := NewSummaryWithTestData()
	s2 := NewSummary()
	samples := []int64{32987, 987, 9879, 879, 87938327, 9823, 25585683826277574}

	for i, v := range samples {
		s2.Insert(v, uint64(42000+i))
	}

	assert.Equal(len(TestArray), s.N)
	s.Merge(s2)
	assert.Equal(len(TestArray)+len(samples), s.N)

	s.Quantile(0.9)
	// FIXME[leo] assert results of merged quantiles
}

func TestSummaryGob(t *testing.T) {
	assert := assert.New(t)

	s := NewSummaryWithTestData()
	bytes, err := s.GobEncode()
	assert.Nil(err)
	ss := NewSummary()
	ss.GobDecode(bytes)

	assert.Equal(s.N, ss.N)
}

func TestSummaryBySlices(t *testing.T) {
	s := NewSummary()

	for i := 0; i < 10000; i++ {
		s.Insert(int64(i), uint64(i))
	}

	slices := s.BySlices(10)
	json.Marshal(slices)
	// FIXME: assert the data, it's not a test!
}

func TestQuantilesMerging(t *testing.T) {
	assert := assert.New(t)

	// summaries to merge
	m1 := NewSummary()
	m2 := NewSummary()

	// an original summary
	o := NewSummary()
	o2 := NewSummary()

	count := 10000
	samples := make([]int64, count)
	for i := 0; i < count; i++ {
		//samples[i] = rand.Int63()
		samples[i] = int64(i)
	}

	// sample a bunch of values
	for _, s := range samples {
		// sample twice into the "original" summary
		o.Insert(s, uint64(s))
		o.Insert(s, uint64(s))
		// sample once into the summaries we'll merge
		m1.Insert(s, uint64(s))
		m2.Insert(s, uint64(s))
	}

	for i := 0; i < 2; i++ {
		for _, s := range samples {
			o2.Insert(s, uint64(s))
		}
	}

	// merge them, make sure we're close.
	m1.Merge(m2)

	for i := 0; i < 11; i++ {
		fmt.Println()

		q := float64(i) / 10.0
		var v int64
		v, _ = o.Quantile(q)
		fmt.Printf("o1 q:%.3f v:%d d:%d\n", q, v, int64(q*10000))

		v, _ = m1.Quantile(q)
		fmt.Printf("m1 q:%.3f v:%d d:%d\n", q, v)

		v, _ = m2.Quantile(q)
		fmt.Printf("m2 q:%.3f v:%d d:%s\n", q, v)

		v, _ = o2.Quantile(q)
		fmt.Printf("o2 q:%.3f v:%d d:%s\n", q, v)
	}

	assert.True(false)

}
