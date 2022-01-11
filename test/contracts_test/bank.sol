pragma solidity ^0.5.8;

contract InvestmentFund {
    uint8 private clientCount;
    mapping (address => uint) private balances;
    address public owner;

  // Логирование событий
    event LogDepositMade(address indexed accountAddress, uint amount);

    // Конструктор контракта
    constructor() public payable {
        require(msg.value == 30 ether, "30 ether initial funding required");
        owner = msg.sender;
        clientCount = 0;
    }

    /// Пополнение баланса
    function deposit() public payable returns (uint) {
        balances[msg.sender] += msg.value;
        emit LogDepositMade(msg.sender, msg.value);
        return balances[msg.sender];
    }

    /// Вывод баланса
    function withdraw(uint withdrawAmount) public returns (uint remainingBal) {
        // Check enough balance available, otherwise just return balance
        if (withdrawAmount <= balances[msg.sender]) {
            balances[msg.sender] -= withdrawAmount;
            msg.sender.transfer(withdrawAmount);
        }
        return balances[msg.sender];
    }

    /// Возвращает баланс аккаунта
    function balance() public view returns (uint) {
        return balances[msg.sender];
    }

    /// Возвращает баланс фонда
    function depositsBalance() public view returns (uint) {
        return address(this).balance;
    }

    /// Вывод средств с фонда на адрес организации
    function pay(address payable organization) public returns (uint remainingBal) {
        if (address(this).balance >= 0 && msg.sender == owner) {
            organization.transfer(address(this).balance);
        }
        return address(this).balance;
    }
}

