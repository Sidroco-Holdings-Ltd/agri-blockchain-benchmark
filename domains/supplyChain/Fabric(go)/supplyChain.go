package chaincode

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Crop struct {
	CropID        string `json:"cropID"`
	Name          string `json:"name"`
	Farmer        string `json:"farmer"`
	CurrentOwner  string `json:"currentOwner"`
	FieldLocation string `json:"fieldLocation"`
	Timestamp     string `json:"timestamp"`
}

type HistoryQueryResult struct {
	Record    *Crop    `json:"record"`
	TxId      string   `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool     `json:"isDelete"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	crops := []Crop{
		{
			CropID:        "C001",
			Name:          "Wheat",
			Farmer:        "Farmer1",
			CurrentOwner:  "Owner1",
			FieldLocation: "Field1",
			Timestamp:     time.Now().String(),
		},
		{
			CropID:        "C002",
			Name:          "Corn",
			Farmer:        "Farmer2",
			CurrentOwner:  "Owner2",
			FieldLocation: "Field2",
			Timestamp:     time.Now().String(),
		},
	}

	for _, crop := range crops {
		cropJSON, err := json.Marshal(crop)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(crop.CropID, cropJSON)
		if err != nil {
			return fmt.Errorf("failed to put crop %s to world state: %v", crop.CropID, err)
		}
	}

	return nil
}

func (s *SmartContract) RegisterCrop(ctx contractapi.TransactionContextInterface, cropID, name, farmer, currentOwner, fieldLocation string) error {
	exists, err := s.CropExists(ctx, cropID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the crop %s already exists", cropID)
	}

	crop := Crop{
		CropID:        cropID,
		Name:          name,
		Farmer:        farmer,
		CurrentOwner:  currentOwner,
		FieldLocation: fieldLocation,
		Timestamp:     time.Now().String(),
	}

	cropJSON, err := json.Marshal(crop)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(cropID, cropJSON)
}

func (s *SmartContract) ReadCrop(ctx contractapi.TransactionContextInterface, cropID string) (*Crop, error) {
	cropJSON, err := ctx.GetStub().GetState(cropID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if cropJSON == nil {
		return nil, fmt.Errorf("the crop %s does not exist", cropID)
	}

	var crop Crop
	err = json.Unmarshal(cropJSON, &crop)
	if err != nil {
		return nil, err
	}

	return &crop, nil
}

func (s *SmartContract) TransferCrop(ctx contractapi.TransactionContextInterface, cropID, newOwner string) error {
	crop, err := s.ReadCrop(ctx, cropID)
	if err != nil {
		return err
	}

	crop.CurrentOwner = newOwner
	crop.Timestamp = time.Now().String()

	cropJSON, err := json.Marshal(crop)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(cropID, cropJSON)
}

func (s *SmartContract) CropExists(ctx contractapi.TransactionContextInterface, cropID string) (bool, error) {
	cropJSON, err := ctx.GetStub().GetState(cropID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return cropJSON != nil, nil
}

func (s *SmartContract) GetAllCrops(ctx contractapi.TransactionContextInterface) ([]*Crop, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var crops []*Crop
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var crop Crop
		err = json.Unmarshal(queryResponse.Value, &crop)
		if err != nil {
			return nil, err
		}
		crops = append(crops, &crop)
	}

	return crops, nil
}

func (s *SmartContract) GetCropHistory(ctx contractapi.TransactionContextInterface, cropID string) ([]HistoryQueryResult, error) {
	log.Printf("GetCropHistory: ID %v", cropID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(cropID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var crop Crop
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &crop)
			if err != nil {
				return nil, err
			}
		} else {
			crop = Crop{
				CropID: cropID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &crop,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}
