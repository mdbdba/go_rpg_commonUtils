package go_rpg_roller

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"encoding/json"
)

type roll struct {
	Options string
	Sides   int
	TimesToRoll int
	RollsGenerated []int
	RollsUsed []int
	Result int
}

func IntSliceToString(src []int) (tgt string) {
	tgt ="["
	for i:=0; i<len(src);i++ {
		joinChr := ", "
		if i == 0 {
			joinChr = ""
		}
		tgt = fmt.Sprintf("%s%s%d", tgt,joinChr,src[i])
	}
	tgt = fmt.Sprintf("%s]", tgt)
	return
}
func (r *roll) ToJson() string {
	j, err := json.Marshal(r)
	if err != nil {
		panic("Issue converting roll to json object")
	}
	return string(j)
}

func (r *roll) ToPrettyString() string {
	return r.ConvertToString(true)
}

func (r *roll) ToString() string {
	return r.ConvertToString(false)
}

func (r *roll) ConvertToString(p bool) (s string) {
	usedStr := IntSliceToString(r.RollsUsed)
	genStr := IntSliceToString(r.RollsGenerated)
	pStr := ""
	if p {
		pStr = "\n\t"
	}
	s = fmt.Sprintf("ROLL -- %sSides: %d, %sTimesToRoll: %d, " +
		"%sOptions: [%s], %sResult: %d, %sRollsUsed: %s, %sRollsGenerated: %s\n",
		pStr, r.Sides,
		pStr, r.TimesToRoll,
		pStr, strings.TrimSpace(r.Options),
		pStr, r.Result,
		pStr, usedStr,
		pStr, genStr)
	return
}

func getRolls(sides int, timesToRoll int) (*[]int, error) {
	var rolls []int
	for i:=0; i< timesToRoll; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
		if err != nil {
			return &[]int{}, err
		}
		t := int(value.Int64()) +1  // +1 because dice start at 1 not 0
		rolls = append(rolls, t )
	}

    return &rolls, nil
}

// Perform - controller that handles getting the result made.
// Options
//   [keep | drop] [highest | lowest] timesToRoll
//   [advantage | disadvantage]
//   [add | subtract] value
//
// Expectations:
//   * advantage and disadvantage cancel each other out.
//     If advantage and disadvantage are both passed, then the
//     result will be a normal roll.
//   * advantage and disadvantage do not stack.
//   * it is assumed that rolling with advantage or disadvantage the
//     number of rolls is = 1. If something other than 1 is passed in this
//     scenario an error will be returned.
//   * using the variadic function for the Options parameter will allow us
//     to simplify all the different combinations by just evaluating them here.
//
func Perform(sides int, timesToRoll int, options ...string ) (r *roll, err error) {
	var reqLogStr string  // boil down all the Options to an easy to read string
	var vantageLogStr string
	var keepLogStr string
	var additiveLogStr string
	keepValue := timesToRoll     // the total number of rolls to keep.
	evalValue := timesToRoll     // the total number of rolls to evaluate.
	   // e.g.  rolling with advantage/disadvantage evals 2 rolls but keep 1
	   //       keep / drop will have eval & keep numbers that differ as well
    sortDirection := "descending"
	additiveValue := 0     // value to add or subtract from the result.
	vantageTrack := "normal"
	for _, opt := range options {
		optSlice := strings.Split(opt, " ")
		switch optSlice[0] {
		case "keep":
			if optSlice[1] == "highest" {
				sortDirection = "descending"
			} else if optSlice[1] == "lowest" {
				sortDirection = "ascending"
			} else {
				panic("Unrecognized string for which values to keep.")
			}
			keepValue, err = strconv.Atoi(optSlice[2])
			if err != nil {
				panic(err)
			}
			keepLogStr = fmt.Sprintf("keep: %d; ",keepValue)
		case "drop":
			if optSlice[1] == "highest" {
				sortDirection = "descending"
			} else if optSlice[1] == "lowest" {
				sortDirection = "ascending"
			} else {
				panic("Unrecognized string for which values to keep.")
			}
			var tmpInt int
			tmpInt, err = strconv.Atoi(optSlice[2])
			if err != nil {
				panic(err)
			}

			keepLogStr = fmt.Sprintf("drop: %d; ",tmpInt)
			if keepValue > tmpInt {
				keepValue -= tmpInt
			} else {
				panic("Tried to drop more rolls than requested.")
			}
		case "add":
			var tValue int
			tValue, err = strconv.Atoi(optSlice[1])
			if err != nil {
				panic(err)
			}
			additiveValue += tValue
			additiveLogStr = fmt.Sprintf("%sadd: %d; ",additiveLogStr,additiveValue)
		case "subtract":
			var tValue int
			tType := "subtract"
			tValue, err = strconv.Atoi(optSlice[1])
			additiveValue -= tValue
			if err != nil {
				panic(err)
			}

			if additiveValue < 0 {
				tType = "add"
			}
			additiveLogStr = fmt.Sprintf("%s%s: %d; ",additiveLogStr,tType, additiveValue)
		case "advantage":
			if timesToRoll != 1 {
				panic("advantage cannot be used with multiple rolls")
			}
			evalValue = 2
			if vantageTrack == "normal" {
				vantageTrack = "advantage"
			} else if vantageTrack == "disadvantage" {
				vantageTrack = "normal"
			}
			vantageLogStr = fmt.Sprintf("vantage: %s; ",vantageTrack)
		case "disadvantage":
			if timesToRoll != 1 {
				panic("disadvantage cannot be used with multiple rolls")
			}
			evalValue = 2
			if vantageTrack == "normal" {
				vantageTrack = "disadvantage"
			} else if vantageTrack == "advantage" {
				vantageTrack = "normal"
			}
			vantageLogStr = fmt.Sprintf("vantage: %s; ",vantageTrack)
		}
	}
	rolls, err :=getRolls(sides,timesToRoll)
    if err != nil {
		panic(err)
	}
	if sortDirection == "ascending" {
		sort.Sort(sort.Reverse(sort.IntSlice(*rolls)))
	} else {
		sort.Ints(*rolls)
	}
	usedSlice := *rolls
	if evalValue != keepValue {
		usedSlice = usedSlice[0:keepValue]
	}
	result := 0
	for i := 0; i < len(usedSlice); i++ {
		result = result + usedSlice[i]
	}
	reqLogStr = fmt.Sprintf("%s%s%s", vantageLogStr, keepLogStr, additiveLogStr)

	return &roll{
		Options:        reqLogStr,
		Sides:          sides,
		TimesToRoll:    timesToRoll,
		RollsGenerated: *rolls,
		RollsUsed:      usedSlice,
		Result:         result,
	}, nil
}

