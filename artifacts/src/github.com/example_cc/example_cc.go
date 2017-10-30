/*
Copyright IBM Corp. 2016 All Rights Reserved.

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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("example_cc0")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var billIndexStr = "_billindex"				//name for the key/value that will store a list of all known bills
var paymentStr = "_paymentindex"	        //name for the key/value that will store a list of all known payments

// Define the Bill structure, with 11 properties.  Structure tags are used by encoding/json library
type Bill struct {
        ID string `json:"id"`
        BillID string `json:"billid"`
        RecipientID string `json:"recipientid"`
        UserID string `json:"userid"`
        FirstName  string `json:"firstname"`
        LastName string `json:"lastname"`
        BillDate string `json:"billdate"`
		BillDueDate string `json:"billduedate"`
		CreatedAt string `json:"created_at"`
		Description string `json:"description"`
        Amount  string `json:"amount"`        
        Currency string `json:"currency"`
        Image string `json:"image"`
        Timestamp string `json:"tr_time"`	//utc timestamp of creation
}

type AllBills struct{
	Bills []Bill `json:"bls"`
}

// Define the Payment structure, with 12 properties.  Structure tags are used by encoding/json library
type Payment struct {
        ID string `json:"id"`
        UserID string `json:"userid"`
        FirstName  string `json:"firstname"`
        LastName string `json:"lastname"`
        Status string `json:"status"`        
        ExchRate string `json:"exchrate"`
        Fees string `json:"fees"`
        FxRate string `json:"fxrate"`
        SourceAmount string `json:"samount"`
		TargetAmount string `json:"tamount"`
		SourceCurrency string `json:"scurrency"`
		TargetCurrency string `json:"tcurrency"`
        Memo string `json:"memo"`        
        ProcessedAt string `json:"processedat"`
        CreatedAt string `json:"createdat"`
        Timestamp string `json:"tr_time"`	//utc timestamp of creation
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### example_cc0 Init ###########")

	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	logger.Info("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}


	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the bill index
	err = stub.PutState(billIndexStr, jsonAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//var empty []string
	//jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the payment index
	err = stub.PutState(paymentStr, jsonAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### example_cc0 Invoke ###########")

	function, args := stub.GetFunctionAndParameters()
	
	if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	}
	if function == "query" {
		// queries an entity state
		return t.query(stub, args)
	}
	if function == "move" {
		// Deletes an entity from its state
		return t.move(stub, args)
	}
    if function == "createBill" {
            return t.createBill(stub, args)
    }
    if function == "queryBill" {
		return t.queryBill(stub, args)
	}
    if function == "createPayment" {
            return t.createPayment(stub, args)
    }
    if function == "queryPayment" {
		return t.queryPayment(stub, args)
	}
    if function == "queryTxsByRange" {
		return t.queryTxsByRange(stub, args)
	}
	if function == "queryBillIDsBasedOnUser" {
	        return t.queryBillIDsBasedOnUser(stub, args)
	}
	if function == "queryBillsBasedOnUser" {
	        return t.queryBillsBasedOnUser(stub, args)
	}
	if function == "queryPaymentsBasedOnUser" {
        return t.queryPaymentsBasedOnUser(stub, args)
	}
	if function == "queryByDate" {
        return t.queryByDate(stub, args)
	}	

	

	logger.Errorf("Unknown action, check the first argument, must be one of 'delete', 'query', 'move', 'createBill', 'queryBill', 'createPayment', 'queryPayment', or 'queryAllBills'. But got: %v", args[0])
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'delete', 'query', 'move', 'createBill', 'queryBill', 'createPayment', 'queryPayment', or 'queryAllBills'. But got: %v", args[0]))
}

func (t *SimpleChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	logger.Infof("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte( .Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

        return shim.Success(nil);
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	logger.Infof("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func (t *SimpleChaincode) createBill(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        if len(args) != 13 {
                return shim.Error("Incorrect number of arguments. Expecting 13")
        }

        billTrTime := time.Now().String()

        var bill = Bill{ID: args[0], BillID: args[1], RecipientID: args[2], UserID: args[3], FirstName: args[4], LastName: args[5], BillDate: args[6], BillDueDate: args[7], CreatedAt: args[8], Description: args[9], Amount: args[10], Currency: args[11], Image: args[12], Timestamp: billTrTime}

        billAsBytes, _ := json.Marshal(bill)
        stub.PutState("BILL"+strconv.Itoa(args[0]), billAsBytes)
        //stub.PutState(args[0], billAsBytes)


		//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
		//  An 'index' is a normal key/value entry in state.
		//  The key is a composite key, with the elements that you want to range query on listed first.
		//  In our case, the composite key is based on indexName~color~name.
		//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
		indexName := "userid~id"
		personIndexKey, err := stub.CreateCompositeKey(indexName, []string{bill.UserID, bill.ID})
		if err != nil {
			return shim.Error(err.Error())
		}
		//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
		//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
		value := []byte{0x00}
		stub.PutState(personIndexKey, value)


		//  ==== Index the bills ====
		//  Write all bill IDs under the predefined string index
		//  ==== Index the bills ====
		
		jsonAsBytes, _ := json.Marshal(bill)
		//get the bills index
		billAsByte , err := stub.GetState(billIndexStr)
		if err != nil {
			return shim.Error("Failed to get bill index")
		}
/*		var billIndex []string
		json.Unmarshal(billAsByte, &billIndex)							//un stringify it aka JSON.parse()
		
		//append
		billIndex = append(billIndex, args[0])									//add points name to index list
		fmt.Println("! bill index: ", billIndex)
		jsonAsBytes, _ := json.Marshal(billIndex)
		err = stub.PutState(billIndexStr, jsonAsBytes) 
		//  ==== End Index the bills ====*/


		var bills AllBills
		json.Unmarshal(billAsByte, &bills)										//un stringify it aka JSON.parse()
		
		bills.Bills = append(bills.Bills, bill);						        //append new bill to Bills
		fmt.Println("! appended new Bill to Bills")
		jsonAsBytes, _ = json.Marshal(bills)
		stub.PutState(billIndexStr, jsonAsBytes)								//rewrite Bills


		// ==== Bill saved and indexed. Return success ====
        return shim.Success(nil)
}

func (t *SimpleChaincode) queryBill(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	billAsBytes, _ := stub.GetState(args[0])
	return shim.Success(billAsBytes)
}

func (t *SimpleChaincode) createPayment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        if len(args) != 15 {
                return shim.Error("Incorrect number of arguments. Expecting 15")
        }

        paymentTrTime := time.Now().String()

        var pay = Payment{ID: args[0], UserID: args[1], FirstName: args[2], LastName: args[3], Status: args[4], ExchRate: args[5], Fees: args[6], FxRate: args[7], SourceAmount: args[8], TargetAmount: args[9], SourceCurrency: args[10], TargetCurrency: args[11], Memo: args[12], ProcessedAt: args[13], CreatedAt: args[14], Timestamp: paymentTrTime}

        payAsBytes, _ := json.Marshal(pay)
        stub.PutState("PAYMENT"+strconv.Itoa(args[0]), payAsBytes)
        stub.PutState(args[0], payAsBytes)


        indexName := "userid~id"
		peymentIndexKey, err := stub.CreateCompositeKey(indexName, []string{pay.UserID, pay.ID})
		if err != nil {
			return shim.Error(err.Error())
		}
		//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
		//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
		value := []byte{0x00}
		stub.PutState(peymentIndexKey, value)

        return shim.Success(nil)
}

func (t *SimpleChaincode) queryPayment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	payAsBytes, _ := stub.GetState(args[0])
	return shim.Success(payAsBytes)
}


// ==== Get Any recorded transaction by indicating range of Keys =========================================
// GetBills By Range, GetPayments By Range, GetTrxs By Range
// ===========================================================================================
func (t *SimpleChaincode) queryTxsByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllBills:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}


// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
//queryBillIDsBasedOnUser will query IDs of Bills  of a given User.
// Uses a GetStateByPartialCompositeKey (range query) against color~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================

func (t *SimpleChaincode) queryBillIDsBasedOnUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    logger.Info("########### queryBillIDsBasedOnUser ###########")
        //   0
        // "userid"
        if len(args) < 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }

        userId := args[0]
        fmt.Println("- start queryBillIDsBasedOnUser ", userId)

        // Query the userid~id index by color
        // This will execute a key range query on all keys starting with 'userid'
        userBillResultsIterator, err := stub.GetStateByPartialCompositeKey("userid~id", []string{userId})
        if err != nil {
                return shim.Error(err.Error())
        }
        defer userBillResultsIterator.Close()

        // Iterate through result set and for each user's bill found, save data
        //var founded AllBills
        var i int
        //response := []string{}
    var buffer bytes.Buffer
        buffer.WriteString("[")
        bArrayMemberAlreadyWritten := false
        for i = 0; userBillResultsIterator.HasNext(); i++ {
                // Note that we don't get the value (2nd return variable), we'll just get the user's bill ID from the composite key
                responseRange, err := userBillResultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }

                // get user and billid from userid~id composite key
                objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
                if err != nil {
                        return shim.Error(err.Error())
                }
                returnedUser := compositeKeyParts[0]
                returnedBillID := compositeKeyParts[1]
                fmt.Printf("- found a Bill from index:%s userid:%s id:%s\n", objectType, returnedUser, returnedBillID)

            // Add a comma before array members, suppress it for the first array member
                if  bArrayMemberAlreadyWritten == true {
                        buffer.WriteString(",")
                }

                buffer.WriteString("{\"Bill\":")
                buffer.WriteString("\"")
                buffer.WriteString("BILL"+string(returnedBillID))
                buffer.WriteString("\"")
                buffer.WriteString("}")
                bArrayMemberAlreadyWritten = true

                //response := []string{returnedUser, returnedBillID}

                //s[i] = returnedBillID
                //founded.Bills = append(founded.Bills, objectType)
                //arr := []byte(returnedBillID)
        }
        buffer.WriteString("]")

        //jsonAsBytes, _ := json.Marshal(founded)

        //fmt.Println("- end queryBillIDsBasedOnUser: ", arr)
        //responsePayload := fmt.Sprintf(returnedBillID)
        //return shim.Success(responsePayload)
        fmt.Printf("- queryBillIDsBasedOnUser:\n%s\n", buffer.String())

        return shim.Success(buffer.Bytes())
}


// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
//queryBillsBasedOnUser will query Bills of a given User.
// Uses a GetStateByPartialCompositeKey (range query) against color~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================

func (t *SimpleChaincode) queryBillsBasedOnUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    logger.Info("########### queryBillsBasedOnUser ###########")
        //   0
        // "userid"
        if len(args) < 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }
        userId := args[0]
        fmt.Println("- start queryBillsBasedOnUser ", userId)

        // Query the userid~id index by color
        // This will execute a key range query on all keys starting with 'userid'
        userBillResultsIterator, err := stub.GetStateByPartialCompositeKey("userid~id", []string{userId})
        if err != nil {
                return shim.Error(err.Error())
        }
        defer userBillResultsIterator.Close()

        // Iterate through result set and for each user's bill found, save data
        var i int
        var s []byte
        for i = 0; userBillResultsIterator.HasNext(); i++ {
                // Note that we don't get the value (2nd return variable), we'll just get the user's bill ID from the composite key
                responseRange, err := userBillResultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }

                // get user and billid from userid~id composite key
                objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
                if err != nil {
                        return shim.Error(err.Error())
                }
                returnedUser := compositeKeyParts[0]
                returnedID := "BILL"+compositeKeyParts[1]
                fmt.Printf("- found a Bill from index:%s userid:%s id:%s\n", objectType, returnedUser, returnedID)

                //billAsBytes, _ := stub.GetState(returnedID)
                //s = append(s, billAsBytes...)
                s = append(s, returnedID...)

        }
        return shim.Success(s)
}


// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
//queryBillsBasedOnUser will query Bills of a given User.
// Uses a GetStateByPartialCompositeKey (range query) against color~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================

func (t *SimpleChaincode) queryPaymentsBasedOnUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    logger.Info("########### queryBillsBasedOnUser ###########")
        //   0
        // "userid"
        if len(args) < 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }
        userId := args[0]
        fmt.Println("- start queryPaymentsBasedOnUser ", userId)

        // Query the userid~id index by color
        // This will execute a key range query on all keys starting with 'userid'
        userPayResultsIterator, err := stub.GetStateByPartialCompositeKey("userid~id", []string{userId})
        if err != nil {
                return shim.Error(err.Error())
        }
        defer userPayResultsIterator.Close()

        // Iterate through result set and for each user's bill found, save data
        var i int
        var s []byte
        for i = 0; userPayResultsIterator.HasNext(); i++ {
                // Note that we don't get the value (2nd return variable), we'll just get the user's bill ID from the composite key
                responseRange, err := userPayResultsIterator.Next()
                if err != nil {
                        return shim.Error(err.Error())
                }

                // get user and billid from userid~id composite key
                objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
                if err != nil {
                        return shim.Error(err.Error())
                }
                returnedUser := compositeKeyParts[0]
                returnedID := compositeKeyParts[1]
                fmt.Printf("- found a Bill from index:%s userid:%s id:%s\n", objectType, returnedUser, returnedID)

                payAsBytes, _ := stub.GetState(returnedID)
                s = append(s, payAsBytes...)

        }
        return shim.Success(s)
}



func inTimeSpan(start, end, check time.Time) bool {
    return check.After(start) && check.Before(end)
}


// ==== QueryByDate =========================================
//queryByDate will query Trx by a given date range.
// ===========================================================================================

func (t *SimpleChaincode) queryByDate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    logger.Info("########### queryByDate ###########")
        //  From           To
        // "2015-10-26"   "2017-11-20"
        if len(args) < 2 {
                return shim.Error("Incorrect number of arguments. Expecting 2")
        }

        fromDate, _ := time.Parse("2006-01-02", args[0])
        toDate, _ := time.Parse("2006-01-02", args[1])
/*        dateType := args[2]

        if (dateType != 'BillDate' || dateType != 'BillDueDate' || dateType != 'CreatedAt') {
        	return shim.Error("Error: Check the first argument. Bill date type must be one of 'BillDate', 'BillDueDate' or 'CreatedAt'.  But got: %v", dateType)
        }*/

		// Check Bill index if it's not empty
		blAsbytes, err := stub.GetState(billIndexStr)	
		if err != nil {
			return shim.Error("Error: Failed to get state")
		}

		// Create a var from Bill structure
		var bills AllBills
		// Read that structure for Transaction Index
		json.Unmarshal(blAsbytes, &bills)

		var founded AllBills
		for i := range bills.Bills{

			bill_date, _ := time.Parse("2006-01-02", bills.Bills[i].BillDueDate)
			fmt.Println("Bill Date a%", bills.Bills[i].BillDueDate)
			if err == nil {}
		    if inTimeSpan(fromDate, toDate, bill_date) {
		        fmt.Println(bill_date, "is between", fromDate, "and", toDate, ".")
		        founded.Bills = append(founded.Bills,bills.Bills[i])
		    }
		}

		jsonAsBytes, _ := json.Marshal(founded)

        return shim.Success(jsonAsBytes)
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
