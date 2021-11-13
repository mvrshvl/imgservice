pragma solidity ^0.8.0;

import "./deps/ERC20.sol";

contract SimpleToken is ERC20 {
  constructor (
    string memory _name,
    string memory _symbol,
    uint256 _amount
  )
    ERC20(_name, _symbol)
    public
  {
    require(_amount > 0, "amount has to be greater than 0");
    _totalSupply = _amount * 10 ** uint256(decimals());
    _balances[msg.sender] = _totalSupply;
    emit Transfer(address(0), msg.sender, _totalSupply);
  }
}