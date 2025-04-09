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
    ID        string `json:"id"`        
    Data      string `json:"data"`      
    Timestamp string `json:"timestamp"` 
}

type CropHistoryQueryResult struct {
    Record    *CropRecord `json:"record"`
    TxId      string      `json:"txId"`
    Timestamp time.Time   `json:"timestamp"`
    IsDelete  bool        `json:"isDelete"`
}

func (f *SmartContract) InitFarm(ctx contractapi.TransactionContextInterface) error {
    crops := []CropRecord{
        {
            ID:        "Crop1",
            Data:      "Initial Crop Data 1",
            Timestamp: time.Now().String(),
        },
        {
            ID:        "Crop2",
            Data:      "Initial Crop Data 2",
            Timestamp: time.Now().String(),
        },
    }

    for _, crop := range crops {
        cropJSON, err := json.Marshal(crop)
        if err != nil {
            return err
        }

        err = ctx.GetStub().PutState(crop.ID, cropJSON)
        if err != nil {
            return fmt.Errorf("failed to store crop record %s in the world state: %v", crop.ID, err)
        }
    }

    return nil
}

func (f *SmartContract) PlantCrop(ctx contractapi.TransactionContextInterface, id string, data string) error {
    exists, err := f.CropExists(ctx, id)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("the crop record %s already exists", id)
    }

    crop := CropRecord{
        ID:        id,
        Data:      data,
        Timestamp: time.Now().String(),
    }

    cropJSON, err := json.Marshal(crop)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(id, cropJSON)
}

func (f *SmartContract) UpdateCrop(ctx contractapi.TransactionContextInterface, id string, data string) error {
    exists, err := f.CropExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("the crop record %s does not exist", id)
    }

    crop := CropRecord{
        ID:        id,
        Data:      data,
        Timestamp: time.Now().String(),
    }

    cropJSON, err := json.Marshal(crop)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(id, cropJSON)
}

func (f *SmartContract) HarvestCrop(ctx contractapi.TransactionContextInterface, id string) (*CropRecord, error) {
    cropJSON, err := ctx.GetStub().GetState(id)
    if err != nil {
        return nil, fmt.Errorf("failed to read crop record from world state: %v", err)
    }
    if cropJSON == nil {
        return nil, fmt.Errorf("the crop record %s does not exist", id)
    }

    var crop CropRecord
    err = json.Unmarshal(cropJSON, &crop)
    if err != nil {
        return nil, err
    }

    return &crop, nil
}

func (f *SmartContract) GetAllCrops(ctx contractapi.TransactionContextInterface) ([]*CropRecord, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var crops []*CropRecord
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var crop CropRecord
        err = json.Unmarshal(queryResponse.Value, &crop)
        if err != nil {
            return nil, err
        }
        crops = append(crops, &crop)
    }

    return crops, nil
}

func (f *SmartContract) GetCropHistory(ctx contractapi.TransactionContextInterface, id string) ([]CropHistoryQueryResult, error) {
    log.Printf("GetCropHistory: ID %v", id)

    resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var crops []CropHistoryQueryResult
    for resultsIterator.HasNext() {
        response, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var crop CropRecord
        if len(response.Value) > 0 {
            err = json.Unmarshal(response.Value, &crop)
            if err != nil {
                return nil, err
            }
        } else {
            crop = CropRecord{
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
            Record:    &crop,
            IsDelete:  response.IsDelete,
        }
        crops = append(crops, historyRecord)
    }

    return crops, nil
}

func (f *SmartContract) CropExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
    crop, err := ctx.GetStub().GetState(id)
    if err != nil {
        return false, fmt.Errorf("failed to read crop record from world state: %v", err)
    }

    return crop != nil, nil
}

func (f *SmartContract) RemoveCrop(ctx contractapi.TransactionContextInterface, id string) error {
    exists, err := f.CropExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("the crop record %s does not exist", id)
    }

    return ctx.GetStub().DelState(id)
}
