package main

type ThresholdBound struct {
	Upper int
	Lower int
}

type TrustValue struct {
	Contributer_Count int
	Commit_count      int
	Metric3_count     int
	Metric4_count     int
	Metric5_count     int
}

type TrustValueThreshold struct {
	Contributer_Count ThresholdBound
	Commit_Count      ThresholdBound
	Metric1           ThresholdBound
	Metric2           ThresholdBound
	Metric3           ThresholdBound
}

func getTrustValueTreshold() {
	/*
		make a magical database request to get populate TrustValueThreshold
		return the value
		The returned value should be something like a global variable, initialized at the beginning, available to the analyzer
		how?
	*/

}

func populateTrustValue() TrustValue {
	var trustValue TrustValue

	return trustValue
}
