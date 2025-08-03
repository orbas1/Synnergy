// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title AI Model Registry
/// @author Synnergy Network
/// @notice Stores metadata references for AI models registered on-chain.
contract AIModelRegistry {
    struct Model {
        string name;            // Human-readable model name
        string metadataURI;     // Off-chain metadata (e.g., IPFS hash)
        address owner;          // Address that registered the model
    }

    /// @notice Mapping of model identifiers to their details.
    mapping(uint256 => Model) private models;
    /// @notice Identifier for the next model to be registered.
    uint256 public nextId;

    /// @notice Emitted when a new model is registered.
    /// @param id Identifier assigned to the model.
    /// @param name Name of the model.
    /// @param metadataURI URI pointing to model metadata.
    /// @param owner Address of the model owner.
    event ModelRegistered(
        uint256 indexed id,
        string name,
        string metadataURI,
        address indexed owner
    );

    /// @dev Thrown when querying a model that does not exist.
    error ModelNotFound();

    /// @notice Register a new AI model.
    /// @param name Human-readable name for the model.
    /// @param metadataURI URI to the model's metadata.
    /// @return id Identifier assigned to the registered model.
    function registerModel(string calldata name, string calldata metadataURI)
        external
        returns (uint256 id)
    {
        id = nextId;
        ++nextId;
        models[id] = Model(name, metadataURI, msg.sender);
        emit ModelRegistered(id, name, metadataURI, msg.sender);
    }

    /// @notice Retrieve details for a model.
    /// @param id Identifier of the model.
    /// @return model The model information.
    function getModel(uint256 id) external view returns (Model memory model) {
        model = models[id];
        if (model.owner == address(0)) revert ModelNotFound();
    }
}
