// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract AgriSupplyChain {
    uint public cropCount = 0;

    // Crop statuses: 
    // - Planted: When a crop is first recorded by a farmer.
    // - Harvested: When the farmer harvests the crop.
    // - Transported: When a transporter moves the crop.
    // - Sold: When the crop reaches the market.
    enum CropStatus { Planted, Harvested, Transported, Sold }

    struct Crop {
        uint id;
        string name;
        string farm;             // Origin or location of the farm.
        string marketLocation;   
        CropStatus status;
        address owner;
        address farmer;
        address transporter;
        address marketAgent;
        uint timestamp;
    }

    mapping(uint => Crop) public crops;

    mapping(address => bool) public farmers;
    mapping(address => bool) public transporters;
    mapping(address => bool) public marketAgents;

    event CropAdded(
        uint id,
        string name,
        string farm,
        string marketLocation,
        CropStatus status,
        address owner,
        address farmer
    );
    event CropStatusUpdated(uint id, CropStatus status, address updater);
    event RolesAssigned(address farmer, address transporter, address marketAgent);

    modifier onlyFarmer() {
        require(farmers[msg.sender], "Not a farmer");
        _;
    }

    modifier onlyTransporter() {
        require(transporters[msg.sender], "Not a transporter");
        _;
    }

    modifier onlyMarketAgent() {
        require(marketAgents[msg.sender], "Not a market agent");
        _;
    }

    function assignRoles(
        address _farmer,
        address _transporter,
        address _marketAgent
    ) public {
        farmers[_farmer] = true;
        transporters[_transporter] = true;
        marketAgents[_marketAgent] = true;
        emit RolesAssigned(_farmer, _transporter, _marketAgent);
    }

    function addCrop(
        string memory _name,
        string memory _farm,
        string memory _marketLocation
    ) public onlyFarmer {
        cropCount++;
        crops[cropCount] = Crop(
            cropCount,
            _name,
            _farm,
            _marketLocation,
            CropStatus.Planted,
            msg.sender,
            msg.sender,
            address(0),
            address(0),
            block.timestamp
        );
        emit CropAdded(cropCount, _name, _farm, _marketLocation, CropStatus.Planted, msg.sender, msg.sender);
    }

    // - Harvested: Can only be updated by a farmer.
    // - Transported: Can only be updated by a transporter.
    // - Sold: Can only be updated by a market agent.
    function updateStatus(uint _id, CropStatus _status) public {
        require(_id > 0 && _id <= cropCount, "Crop does not exist.");
        Crop storage crop = crops[_id];

        if (_status == CropStatus.Harvested) {
            require(farmers[msg.sender], "Only farmers can update to Harvested");
            // (Optional) You might update crop.farmer = msg.sender if needed.
        } else if (_status == CropStatus.Transported) {
            require(transporters[msg.sender], "Only transporters can update to Transported");
            crop.transporter = msg.sender;
        } else if (_status == CropStatus.Sold) {
            require(marketAgents[msg.sender], "Only market agents can update to Sold");
            crop.marketAgent = msg.sender;
        } else {
            revert("Invalid status transition");
        }

        crop.status = _status;
        crop.timestamp = block.timestamp;
        emit CropStatusUpdated(_id, _status, msg.sender);
    }

    function getCrop(uint _id) public view returns (
        uint,
        string memory,
        string memory,
        string memory,
        CropStatus,
        address,
        address,
        address,
        address,
        uint
    ) {
        require(_id > 0 && _id <= cropCount, "Crop does not exist.");
        Crop memory crop = crops[_id];
        return (
            crop.id,
            crop.name,
            crop.farm,
            crop.marketLocation,
            crop.status,
            crop.owner,
            crop.farmer,
            crop.transporter,
            crop.marketAgent,
            crop.timestamp
        );
    }
}
