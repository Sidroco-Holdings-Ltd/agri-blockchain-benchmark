// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract DeFinance {
    uint public loanCount = 0;

    enum LoanStatus { Requested, Approved, Rejected, Repaid }

    struct Loan {
        uint id;
        address farmer;
        address lender;
        uint amount; // in wei
        uint interestRate; // in basis points (e.g., 500 = 5%)
        uint duration; // in seconds
        LoanStatus status;
        uint timestamp;
    }

    mapping(uint => Loan) public loans;

    event LoanRequested(uint id, address farmer, uint amount, uint interestRate, uint duration);
    event LoanApproved(uint id, address lender);
    event LoanRejected(uint id, address lender);
    event LoanRepaid(uint id, address farmer);

    /// @notice Request a new loan without sending ETH.
    /// @param _amount The loan amount in wei.
    /// @param _interestRate The interest rate in basis points (e.g., 500 = 5%).
    /// @param _duration The loan duration in seconds.
    function requestLoan(uint _amount, uint _interestRate, uint _duration) public {
        require(_amount > 0, "Loan amount must be greater than zero");
        require(_interestRate > 0, "Interest rate must be greater than zero");
        require(_duration > 0, "Duration must be greater than zero");

        loanCount++;
        loans[loanCount] = Loan(
            loanCount,
            msg.sender,
            address(0), // No lender assigned yet
            _amount,
            _interestRate,
            _duration,
            LoanStatus.Requested,
            block.timestamp
        );

        emit LoanRequested(loanCount, msg.sender, _amount, _interestRate, _duration);
    }

    /// @notice Approve a loan by providing the required ETH.
    /// @param _id The ID of the loan to approve.
    function approveLoan(uint _id) public payable {
        Loan storage loan = loans[_id];
        require(loan.id != 0, "Loan does not exist");
        require(loan.status == LoanStatus.Requested, "Loan is not in a requested state");
        require(msg.value == loan.amount, "Incorrect ETH amount sent for loan approval");

        loan.status = LoanStatus.Approved;
        loan.lender = msg.sender;

        // Transfer the loan amount to the farmer
        payable(loan.farmer).transfer(loan.amount);

        emit LoanApproved(_id, msg.sender);
    }

    /// @notice Reject a loan and refund any ETH if necessary.
    /// @param _id The ID of the loan to reject.
    function rejectLoan(uint _id) public {
        Loan storage loan = loans[_id];
        require(loan.id != 0, "Loan does not exist");
        require(loan.status == LoanStatus.Requested, "Loan is not in a requested state");

        loan.status = LoanStatus.Rejected;

        emit LoanRejected(_id, msg.sender);
    }

    /// @notice Repay an approved loan along with interest.
    /// @param _id The ID of the loan to repay.
    function repayLoan(uint _id) public payable {
        Loan storage loan = loans[_id];
        require(loan.id != 0, "Loan does not exist");
        require(loan.status == LoanStatus.Approved, "Loan is not in an approved state");
        require(msg.sender == loan.farmer, "Only the farmer can repay the loan");

        uint totalRepayment = loan.amount + ((loan.amount * loan.interestRate) / 10000);
        require(msg.value == totalRepayment, "Incorrect repayment amount");

        loan.status = LoanStatus.Repaid;

        // Transfer the repayment to the lender
        payable(loan.lender).transfer(msg.value);

        emit LoanRepaid(_id, msg.sender);
    }
    
    function getLoan(uint _id) public view returns (
        uint id,
        address farmer,
        address lender,
        uint amount,
        uint interestRate,
        LoanStatus status,
        uint timestamp
    ) {
        Loan memory loan = loans[_id];
        require(loan.id != 0, "Loan does not exist");

        return (
            loan.id,
            loan.farmer,
            loan.lender,
            loan.amount,
            loan.interestRate,
            loan.status,
            loan.timestamp
        );
    }
}
