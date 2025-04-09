// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract AgriProcedureManager {
    uint public procedureCount = 0;

    struct ProcedureRecord {
        uint id;
        uint farmId;
        string procedureHash; // Reference to off-chain procedure details (e.g., planting, watering, harvesting)
        string category;      // e.g., Crop Management, Irrigation, Fertilization
        uint timestamp;
        address recorder;
    }

    mapping(uint => ProcedureRecord) public procedures;
    mapping(uint => address[]) public procedureAccessList; 

    event ProcedureRecorded(uint id, uint farmId, string procedureHash, string category, uint timestamp, address recorder);
    event ProcedureAccessGranted(uint id, address grantedTo);
    event ProcedureAccessRevoked(uint id, address revokedFrom);

    modifier procedureExists(uint _id) {
        require(_id > 0 && _id <= procedureCount, "Procedure record does not exist");
        _;
    }

    function recordProcedure(uint _farmId, string memory _procedureHash, string memory _category) public {
        procedureCount++;
        procedures[procedureCount] = ProcedureRecord(
            procedureCount,
            _farmId,
            _procedureHash,
            _category,
            block.timestamp,
            msg.sender
        );
        emit ProcedureRecorded(procedureCount, _farmId, _procedureHash, _category, block.timestamp, msg.sender);
    }

    function grantAccess(uint _id, address _grantee) public procedureExists(_id) {
        ProcedureRecord memory record = procedures[_id];
        require(record.recorder == msg.sender, "Only the recorder can grant access");
        procedureAccessList[_id].push(_grantee);
        emit ProcedureAccessGranted(_id, _grantee);
    }

    function revokeAccess(uint _id, address _revokee) public procedureExists(_id) {
        ProcedureRecord memory record = procedures[_id];
        require(record.recorder == msg.sender, "Only the recorder can revoke access");
        address[] storage accessList = procedureAccessList[_id];
        for(uint i = 0; i < accessList.length; i++) {
            if(accessList[i] == _revokee) {
                accessList[i] = accessList[accessList.length - 1];
                accessList.pop();
                emit ProcedureAccessRevoked(_id, _revokee);
                break;
            }
        }
    }

    function hasAccess(uint _id, address _user) public view procedureExists(_id) returns (bool) {
        if(procedures[_id].recorder == _user) {
            return true;
        }
        address[] memory accessList = procedureAccessList[_id];
        for(uint i = 0; i < accessList.length; i++) {
            if(accessList[i] == _user) {
                return true;
            }
        }
        return false;
    }

    function getProcedure(uint _id) public view procedureExists(_id) returns (
        uint,
        uint,
        string memory,
        string memory,
        uint,
        address
    ) {
        require(hasAccess(_id, msg.sender), "Access denied");
        ProcedureRecord memory record = procedures[_id];
        return (
            record.id,
            record.farmId,
            record.procedureHash,
            record.category,
            record.timestamp,
            record.recorder
        );
    }

    function getFarmProcedures(uint _farmId) public view returns (ProcedureRecord[] memory) {
        uint count = 0;
        for(uint i = 1; i <= procedureCount; i++) {
            if(procedures[i].farmId == _farmId && hasAccess(i, msg.sender)) {
                count++;
            }
        }

        ProcedureRecord[] memory farmProcedures = new ProcedureRecord[](count);
        uint index = 0;
        for(uint i = 1; i <= procedureCount; i++) {
            if(procedures[i].farmId == _farmId && hasAccess(i, msg.sender)) {
                farmProcedures[index] = procedures[i];
                index++;
            }
        }
        return farmProcedures;
    }
}
