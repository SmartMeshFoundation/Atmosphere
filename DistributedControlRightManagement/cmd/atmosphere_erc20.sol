// Abstract contract for the full ERC 20 Token standard
// https://github.com/ethereum/EIPs/issues/20
pragma solidity ^0.4.24;

contract Token {
    /* This is a slight change to the ERC20 base standard.*/
    /// total amount of tokens
    uint256 public totalSupply;

    /// @param _owner The address from which the balance will be retrieved
    /// @return The balance
    function balanceOf(address _owner) public view returns (uint256 balance);

    /// @notice send `_value` token to `_to` from `msg.sender`
    /// @param _to The address of the recipient
    /// @param _value The amount of token to be transferred
    /// @return Whether the transfer was successful or not
    function transfer(address _to, uint256 _value) public returns (bool success);

    /// @notice send `_value` token to `_to` from `_from` on the condition it is approved by `_from`
    /// @param _from The address of the sender
    /// @param _to The address of the recipient
    /// @param _value The amount of token to be transferred
    /// @return Whether the transfer was successful or not
    function transferFrom(address _from, address _to, uint256 _value) public returns (bool success);

    /// @notice `msg.sender` approves `_spender` to spend `_value` tokens
    /// @param _spender The address of the account able to transfer the tokens
    /// @param _value The amount of tokens to be approved for transfer
    /// @return Whether the approval was successful or not
    function approve(address _spender, uint256 _value) public returns (bool success);

    /// @param _owner The address of the account owning tokens
    /// @param _spender The address of the account able to transfer the tokens
    /// @return Amount of remaining tokens allowed to spent
    function allowance(address _owner, address _spender) public view returns (uint256 remaining);

    event Transfer(address indexed _from, address indexed _to, uint256 _value);
    event Approval(address indexed _owner, address indexed _spender, uint256 _value);
}

contract Owned {

    /// `owner` is the only address that can call a function with this
    /// modifier
    modifier onlyOwner() {
        require(msg.sender == owner);
        _;
    }

    address public owner;

    /// @notice The Constructor assigns the message sender to be `owner`
    function Owned() public {
        owner = msg.sender;
    }

    address newOwner=0x0;

    event OwnerUpdate(address _prevOwner, address _newOwner);

    ///change the owner
    function changeOwner(address _newOwner) public onlyOwner {
        require(_newOwner != owner);
        newOwner = _newOwner;
    }

    /// accept the ownership
    function acceptOwnership() public{
        require(msg.sender == newOwner);
        emit OwnerUpdate(owner, newOwner);
        owner = newOwner;
        newOwner = 0x0;
    }
}

contract StandardToken is Token,Owned {

    function transfer(address _to, uint256 _value) public  returns (bool success) {
        //Default assumes totalSupply can't be over max (2^256 - 1).
        //If your token leaves out totalSupply and can issue more tokens as time goes on, you need to check if it doesn't wrap.
        //Replace the if with this one instead.
        if (balances[msg.sender] >= _value && balances[_to] + _value > balances[_to]) {
            balances[msg.sender] -= _value;
            balances[_to] += _value;
            emit Transfer(msg.sender, _to, _value);
            return true;
        } else { return false; }
    }

    function transferFrom(address _from, address _to, uint256 _value) public returns (bool success) {
        //same as above. Replace this line with the following if you want to protect against wrapping uints.
        if (balances[_from] >= _value && allowed[_from][msg.sender] >= _value && balances[_to] + _value > balances[_to]) {
            balances[_to] += _value;
            balances[_from] -= _value;
            allowed[_from][msg.sender] -= _value;
            emit Transfer(_from, _to, _value);
            return true;
        } else { return false; }
    }

    function balanceOf(address _owner) public view returns (uint256 balance) {
        return balances[_owner];
    }

    function approve(address _spender, uint256 _value) public returns (bool success) {
        allowed[msg.sender][_spender] = _value;
        emit Approval(msg.sender, _spender, _value);
        return true;
    }

    function allowance(address _owner, address _spender) public view returns (uint256 remaining) {
        return allowed[_owner][_spender];
    }

    mapping (address => uint256) balances;
    mapping (address => mapping (address => uint256)) allowed;
}

contract EthereumToken is StandardToken {

    function () public {
        revert();
    }

    string public name = "Ethereum Token for atmosphere";                   //fancy name
    uint8 public decimals = 18;                //How many decimals to show. ie. There could 1000 base units with 3 decimals. Meaning 0.980 SBX = 980 base units. It's like comparing 1 wei to 1 ether.
    string public symbol = "EthereumToken";                 //An identifier
    string public version = 'v0.1';       //SMT 0.1 standard. Just an arbitrary versioning scheme.

    mapping (address => uint256) locked; //锁定金额
    mapping (address=> bytes32) main_chain_data; //锁定这部分金额转给
    // The nonce for avoid transfer replay attacks
    mapping(address => uint256) nonces;

    function EthereumToken() public {
        totalSupply=1; //保证total supply大于等于0,符合ERC20规范
    }
//公证人为
    function lockedInForAccount(address account,uint256 value) onlyOwner  public {
        uint256 oldValue=balances[account];
        require(oldValue+value>=value);
        balances[account]=oldValue+value;
        totalSupply=totalSupply+value;
        require(totalSupply>value); // 部署合约时的那个1一直存在,还不能溢出
    }
    event PrePareLockedOut(address indexed _from, uint256 _value);
    //准备退出,需要在合约里记录,公证人需要监控这里的事件,采用相应的操作, 后续应该提供PrePareLockedOutProxy函数,可以保证交易方没有侧链主币的情况下,仍然可以发起合约
    function prePareLockedOut(uint256 value,bytes32 data) public{
        require(locked[msg.sender]==0); // 没有正在退出的历史交易 data=hash(tx)
        require(balances[msg.sender]>=value);
        locked[msg.sender]=value;
        balances[msg.sender]-=value;
        main_chain_data[msg.sender]=data;
        emit PrePareLockedOut(msg.sender,value);
    }
    //公证人在主链上转账到具体地址后,销毁相应记录
    function lockedOut(address from) onlyOwner public {
        require(locked[from]>0);
        locked[from]=0;
        main_chain_data[from]=bytes32(0);
    }
}