package mapreduce

import "hash/fnv"

// doMap does the job of a map worker: it reads the one of the input files
// (inFile), calls the user-defined map function (mapF) for that file's
// contents, and writes the output into nReduce intermediate bins.
func doMap(
	jobName string,
	job int,
	inFile string,
	nReduce int,
	mapF func(string, string) []KeyValue,
) {
	// TODO:
	// You will need to write this function.
	// You can find the intermediate bin filename for reduce job r using
	// reduceName(jobName, job, r). The ihash function (given below doMap) should
	// be used to decide which bin a given key belongs into.
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
