// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract AgriMonitoring {
    uint public dataCount = 0;

    struct SensorData {
        uint id;
        uint farmId;
        string dataType; // e.g., Temperature, Humidity, Soil Moisture
        string dataValue;
        uint timestamp;
        address uploader;
    }

    mapping(uint => SensorData) public sensorDataRecords;

    event DataRecorded(uint id, uint farmId, string dataType, string dataValue, uint timestamp, address uploader);
    event DataUpdated(uint id, string dataValue, uint timestamp, address updater);

    function recordData(uint _farmId, string memory _dataType, string memory _dataValue) public {
        dataCount++;
        sensorDataRecords[dataCount] = SensorData(
            dataCount,
            _farmId,
            _dataType,
            _dataValue,
            block.timestamp,
            msg.sender
        );
        emit DataRecorded(dataCount, _farmId, _dataType, _dataValue, block.timestamp, msg.sender);
    }

    function updateData(uint _id, string memory _newDataValue) public {
        require(_id > 0 && _id <= dataCount, "Data record does not exist");
        SensorData storage record = sensorDataRecords[_id];
        require(record.uploader == msg.sender, "Only uploader can update data");
        record.dataValue = _newDataValue;
        record.timestamp = block.timestamp;
        emit DataUpdated(_id, _newDataValue, block.timestamp, msg.sender);
    }

    function getSensorData(uint _id) public view returns (
        uint,
        uint,
        string memory,
        string memory,
        uint,
        address
    ) {
        require(_id > 0 && _id <= dataCount, "Data record does not exist");
        SensorData memory record = sensorDataRecords[_id];
        return (
            record.id,
            record.farmId,
            record.dataType,
            record.dataValue,
            record.timestamp,
            record.uploader
        );
    }

    function getFarmData(uint _farmId) public view returns (SensorData[] memory) {
        uint count = 0;
        for(uint i = 1; i <= dataCount; i++) {
            if(sensorDataRecords[i].farmId == _farmId) {
                count++;
            }
        }

        SensorData[] memory farmData = new SensorData[](count);
        uint index = 0;
        for(uint i = 1; i <= dataCount; i++) {
            if(sensorDataRecords[i].farmId == _farmId) {
                farmData[index] = sensorDataRecords[i];
                index++;
            }
        }
        return farmData;
    }
}
