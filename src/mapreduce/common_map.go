package mapreduce

import "hash/fnv"

// doMap reads the assigned input file, calls mapF for that file's contents,
// and writes the output into nReduce output bins.
func doMap(
	jobName string,
	job int,
	inFile string,
	nReduce int,
	mapF func(string, string) []KeyValue,
) {
	// TODO:
	// You will need to write this function.
	// You can find the output bin filename for reduce job r using reduceName(jobName, job, r).
	// The ihash function (given below doMap) should be used for output binning.
	//
	// You may choose how you would like to encode the key/value pairs in
	// the intermediate files. One way would be to encode them using JSON.
	// Since the reduce jobs have to output JSON anyway, you should
	// probably do that here too. You can write JSON to a file using
	//
	//   enc := json.NewEncoder(file)
	//   for _, kv := ... {
	//     err := enc.Encode(&kv)
	//
	// Remember to close the file after you have written all the values!
}

func ihash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
