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

type CropRecord struct {
    ID        string  `json:"id"`        
    CropType  string  `json:"cropType"` 
    Yield     float64 `json:"yield"`    
    Timestamp string  `json:"timestamp"` 
}

type CropHistoryQueryResult struct {
    Record    *CropRecord `json:"record"`
    TxId      string      `json:"txId"`
    Timestamp time.Time   `json:"timestamp"`
    IsDelete  bool        `json:"isDelete"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    records := []CropRecord{
        {
            ID:        "Crop1",
            CropType:  "Wheat",
            Yield:     150.5,
            Timestamp: time.Now().String(),
        },
        {
            ID:        "Crop2",
            CropType:  "Corn",
            Yield:     200.2,
            Timestamp: time.Now().String(),
        },
    }

    for _, record := range records {
        recordJSON, err := json.Marshal(record)
        if err != nil {
            return err
        }

        err = ctx.GetStub().PutState(record.ID, recordJSON)
        if err != nil {
            return fmt.Errorf("failed to put crop record %s to world state: %v", record.ID, err)
        }
    }

    return nil
}

func (s *SmartContract) AddCropRecord(ctx contractapi.TransactionContextInterface, id string, cropType string, yield float64) error {
    exists, err := s.CropRecordExists(ctx, id)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("the crop record %s already exists", id)
    }

    record := CropRecord{
        ID:        id,
        CropType:  cropType,
        Yield:     yield,
        Timestamp: time.Now().String(),
    }

    recordJSON, err := json.Marshal(record)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(id, recordJSON)
}

func (s *SmartContract) UpdateCropRecord(ctx contractapi.TransactionContextInterface, id string, cropType string, yield float64) error {
    exists, err := s.CropRecordExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("the crop record %s does not exist", id)
    }

    record := CropRecord{
        ID:        id,
        CropType:  cropType,
        Yield:     yield,
        Timestamp: time.Now().String(),
    }

    recordJSON, err := json.Marshal(record)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(id, recordJSON)
}

func (s *SmartContract) GetCropRecord(ctx contractapi.TransactionContextInterface, id string) (*CropRecord, error) {
    recordJSON, err := ctx.GetStub().GetState(id)
    if err != nil {
        return nil, fmt.Errorf("failed to read crop record from world state: %v", err)
    }
    if recordJSON == nil {
        return nil, fmt.Errorf("the crop record %s does not exist", id)
    }

    var record CropRecord
    err = json.Unmarshal(recordJSON, &record)
    if err != nil {
        return nil, err
    }

    return &record, nil
}

func (s *SmartContract) GetAllCropRecords(ctx contractapi.TransactionContextInterface) ([]*CropRecord, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var records []*CropRecord
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var record CropRecord
        err = json.Unmarshal(queryResponse.Value, &record)
        if err != nil {
            return nil, err
        }
        records = append(records, &record)
    }

    return records, nil
}

func (s *SmartContract) GetCropRecordHistory(ctx contractapi.TransactionContextInterface, id string) ([]CropHistoryQueryResult, error) {
    log.Printf("GetCropRecordHistory: ID %v", id)

    resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var records []CropHistoryQueryResult
    for resultsIterator.HasNext() {
        response, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var record CropRecord
        if len(response.Value) > 0 {
            err = json.Unmarshal(response.Value, &record)
            if err != nil {
                return nil, err
            }
        } else {
            record = CropRecord{
                ID: id,
            }
        }

        timestamp, err := ptypes.Timestamp(response.Timestamp)
        if err != nil {
            return nil, err
        }

        historyRecord := CropHistoryQueryResult{
            TxId:      response.TxId,
            Timestamp: timestamp,
            Record:    &record,
            IsDelete:  response.IsDelete,
        }
        records = append(records, historyRecord)
    }

    return records, nil
}

func (s *SmartContract) CropRecordExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
    record, err := ctx.GetStub().GetState(id)
    if err != nil {
        return false, fmt.Errorf("failed to read crop record from world state: %v", err)
    }
    return record != nil, nil
}

func (s *SmartContract) DeleteCropRecord(ctx contractapi.TransactionContextInterface, id string) error {
    exists, err := s.CropRecordExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("the crop record %s does not exist", id)
    }
    return ctx.GetStub().DelState(id)
}
