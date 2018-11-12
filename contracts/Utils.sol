pragma solidity ^0.4.25;

/// @title Utility Contract
/// @notice a general set of utility functions included into this contract.
contract Utils {
    string constant public contract_version = "0.5._";

    /// @notice Check if a contract exists
    /// @param contract_address The address to check whether a contract is deployed or not
    /// @return True if a contract exists, false otherwise
    function contractExists(address contract_address) public view returns (bool) {
        uint size;

        assembly {
            size := extcodesize(contract_address)
        }

        return size > 0;
    }
}
