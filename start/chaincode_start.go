//chaincode for simple referendum vote election

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type districtReferendum struct {
	DistrictName string
	NoVotes      int
	YesVotes     int
	Votes        map[string]string //maps vote ID to vote
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() { //main function executes when each peer deploys its instance of the chaincode
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting at least one district")
	}

	//create data model
	electionMetaData := &districtReferendum{DistrictName: args[0], NoVotes: 0, YesVotes: 0, Votes: make(map[string]string)} //golang struct
	electionMetaDataJSON, err := json.Marshal(electionMetaData)                                                             //golang JSON (byte array)
	if err != nil {
		return nil, errors.New("Marshalling has failed")
	}

	err = stub.PutState(args[0], electionMetaDataJSON) //writes the key-value pair (args[0] (district name), json object)
	if err != nil {
		return nil, errors.New("put state has failed")
	}

	return nil, nil
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "error" {
		return t.error(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return nil, errors.New("Received unknown function invocation")
}

func (t *SimpleChaincode) error(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	return nil, errors.New("generic error")
}

func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var name string
	var value string
	var err error

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]
	value = args[1]

	//TODO: check if this user has already voted
	//if user has already voted, return with message
	//else, proceed to vote

	preExistVote, err := stub.GetState(name) //gets value for the given key
	if err != nil {
		return nil, err
	}
	if preExistVote != nil { //vote already exists
		return nil, errors.New("vote already exists")
	}

	//if person has not already voted
	metadataRaw, err := stub.GetState("electionMetaData")
	if err != nil { //get state error
		return nil, err
	}

	var metaDataStructToUpdate districtReferendum
	err = json.Unmarshal(metadataRaw, &metaDataStructToUpdate)
	if err != nil { //unmarshalling error
		return nil, err
	}

	if strings.TrimRight(value, "\n") == "yes" {
		metaDataStructToUpdate.YesVotes++

	} else if strings.TrimRight(value, "\n") == "no" {
		metaDataStructToUpdate.NoVotes++
	}

	electionMetaDataJSON, err := json.Marshal(metaDataStructToUpdate) //golang JSON (byte array)
	if err != nil {                                                   //marshall error
		return nil, err
	}

	err = stub.PutState("electionMetaData", electionMetaDataJSON) //writes the key-value pair (electionMetaData, json object)
	if err != nil {
		return nil, err
	}

	err = stub.PutState(name, []byte(value)) //JOSE: writes a key-value pair with the given key and value paramenters. We need to introduce a more complex data model that includes an increasing vote ID, for iterating over votes.
	if err != nil {                          //putstate error
		return nil, err
	}

	return nil, nil
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "dummy_query" { //read a variable
		fmt.Println("hi there " + function) //error
		return nil, nil
	} else if function == "read" {
		return t.read(stub, args)
	} else if function == "error" {
		return t.error(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name string
	var jsonResp string

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //gets value for the given key
	if err != nil {                        //getstate error
		return nil, err
	}

	if valAsbytes == nil { //vote doesn't exist
		jsonResp = "{\"Error\":\"Failed to get vote for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}
