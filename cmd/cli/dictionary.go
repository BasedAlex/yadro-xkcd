package main

func dictionary() map[string]bool {
	filterMap := make(map[string]bool)
	filterMap["i"] = true
	filterMap["i'm"] = true
	filterMap["i'"] = true
	filterMap["i'll"] = true
	filterMap["you"] = true
	filterMap["as"] = true
	filterMap["are"] = true
	filterMap["me"] = true
	filterMap["a"] = true
	filterMap["an"] = true
	filterMap["he"] = true
	filterMap["she"] = true
	filterMap["it"] = true
	filterMap["it's"] = true
	filterMap["there"] = true
	filterMap["they"] = true
	filterMap["they're"] = true
	filterMap["this"] = true
	filterMap["mine"] = true
	filterMap["your"] = true
	filterMap["yours"] = true
	filterMap["will"] = true
	filterMap["did"] = true
	filterMap["does"] = true
	filterMap["could've"] = true
	filterMap["would've"] = true
	filterMap["of"] = true

	return filterMap
}
