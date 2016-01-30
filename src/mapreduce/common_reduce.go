package mapreduce

// doReduce reads the map outputs for the bin assigned to this job, sorts the
// outputs by key, calls reduceF for each key, and writes the output to disk.
func doReduce(
	jobName string,
	job int,
	nMap int,
	reduceF func(string, []string) string,
) {
	// TODO:
	// You will need to write this function.
	// You can find the output bin for this reduce job from map job m using reduceName(jobName, m, job).
	// Remember that the values in the input files are encoded, so you will
	// need to decode them. If you chose to use JSON, you can read out
	// multiple decoded values by creating a decoder, and then repeatedly
	// calling .Decode() on it until Decode() returns an error..
	//
	// You should write the reduced output in as JSON encoded KeyValue
	// objects to a file named mergeName(jobName, job). We require you to
	// use JSON here because that is what the merger than combines the
	// output from all the reduce jobs expects. There is nothing "special"
	// about JSON -- it is just the marshalling format we chose to use. It
	// will look something like this:
	//
	// enc := json.NewEncoder(mergeFile)
	// for key in ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
}
