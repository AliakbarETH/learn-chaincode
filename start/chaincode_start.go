/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
    "strconv"
    "encoding/json"
    "runtime"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {

}

//name for the key/value that will store a list of all known journals
var journalIndexStr = "_journalindex" 

type Journal struct
{
    Name string `json:"name"`
    CPR string `json:"cpr_nr"`
    Status string `json:"status"`
    State string `json:"state"`
    Timestamp string `json:"timestamp"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {

    // maximize CPU usage for maximum performance
    runtime.GOMAXPROCS(runtime.NumCPU())

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - Reset all things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    
    var Aval int
    var err error

    if len(args) != 1 {
        return nil, errors.New("Incorrect number of arguments. Expecting 1")
    }

    // Initialize the chaincode
    Aval, err = strconv.Atoi(args[0])
    if err != nil {
        return nil, errors.New("Expecting integer value for asset holding")
    }

    // Write the state to the ledger
    err = stub.PutState("abc", []byte(strconv.Itoa(Aval))) //making a test var "abc"
    if err != nil {
        return nil, err
    }

    var empty []string
    jsonAsBytes, _ := json.Marshal(empty) //marshal an emtpy array of strings to clear the index
    err = stub.PutState(journalIndexStr, jsonAsBytes)
    if err != nil {
        return nil, err
    }

    return nil, nil
}

// ============================================================================================================================
// Invoke - Our entry point to invoke a chaincode function
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    fmt.Println("invoke is running " + function)

    // Handle different functions
    if function == "init" {
        return t.Init(stub, "init", args)
    } else if function == "write" {
        return t.Write(stub, args)
    } else if function == "init_journal" {   
        //create a new journal                              
        return t.Init_journal(stub, args)
    }
    
    fmt.Println("invoke did not find func: " + function)

    return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    fmt.Println("query is running " + function)

    // Handle different functions
    if function == "read" {                            //read a variable
        return t.Read(stub, args)
    }
    
    fmt.Println("query did not find func: " + function)

    return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    var cpr, value string
    var err error
    fmt.Println("running write()")

    if len(args) != 2 {
        return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
    }

    cpr = args[0]                            //rename for fun
    value = args[1]
    err = stub.PutState(cpr, []byte(value))  //write the variable into the chaincode state
    if err != nil { return nil, err }

    return nil, nil
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    var cpr, jsonResp string
    var err error

    if len(args) != 1 {
    return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
    }

    cpr = args[0]
    valAsbytes, err := stub.GetState(cpr)
    if err != nil {
        jsonResp = "{\"Error\":\"Failed to get state for " + cpr + "\"}"
        return nil, errors.New(jsonResp)
    }

    return valAsbytes, nil
}


// ============================================================================================================================
// Init Journal - create a new journal, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Init_journal(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    var err error

    /* 
    Our model looks like
    -------------------------------------------------------
    Name string, CPR string, Status string, State int64 , Timestamp int64
    -------------------------------------------------------
       0        1        2         3          4
    "name", "cpr-nr", "status", "state", "timestamp"
    -------------------------------------------------------
    */
    
    if len(args) != 5 {
        return nil, errors.New("Incorrect number of arguments. Expecting 5")
    }

    // Input sanitation
    fmt.Println("- start init journal")
    if len(args[0]) <= 0 {
        return nil, errors.New("1st argument must be a non-empty string")
    }
    if len(args[1]) <= 0 {
        return nil, errors.New("2nd argument must be a non-empty string")
    }
    if len(args[2]) <= 0 {
        return nil, errors.New("3rd argument must be a non-empty string")
    }
    if len(args[3]) <= 0 {
        return nil, errors.New("4rd argument must be a non-empty string")
    }
    if len(args[4]) <= 0 {
        return nil, errors.New("5th argument must be a non-empty string")
    }

    // Retrive values
    name := args[0]
    cpr := args[1]
    status := args[2]
    state := args[3]
    timestamp := args[4]

    //check if cpr-nr already exists
    journalAsBytes, err := stub.GetState(cpr)
    if err != nil {
        return nil, errors.New("Failed to get cpr-nr")
    }
    
    res := Journal{}
    json.Unmarshal(journalAsBytes, &res)
    if res.CPR == cpr{
        fmt.Println("This cpr-nr arleady exists: " + cpr)
        fmt.Println(res);
        //all stop a journal if this cpr-nr exists
        return nil, errors.New("This cpr-nr arleady exists")                
    }
    
    //build the journal json string manually
    strJson := `{"name": "` + name + `", "cpr_nr": "` + cpr +  `", "status": ` + status +  `, "state": "` + state + `, "timestamp": "` + timestamp + `"}`
    
    //store journal with cpr as key
    err = stub.PutState(cpr, []byte(strJson))                                  
    if err != nil { return nil, err }
        
    //get the journal index
    journalsAsBytes, err := stub.GetState(journalIndexStr)
    if err != nil { return nil, errors.New("Failed to get journal index") }
    
    var journalIndex []string
    //un stringify it aka JSON.parse()
    json.Unmarshal(journalsAsBytes, &journalIndex)                            
    
    //append - add journal cpr-nr to index list
    journalIndex = append(journalIndex, cpr)                                 
    fmt.Println("! journal index: ", journalIndex)

    //store cpr-nr of journal
    jsonAsBytes, _ := json.Marshal(journalIndex)
    err = stub.PutState(journalIndexStr, jsonAsBytes)                        

    fmt.Println("- end init journal")
    return nil, nil
}



