# Photon Glossary

- `counterparty` :The counterparty of a channel is the other participant of the channel that is not ourselves.
- `settlement window` : The number of blocks after the closing of a channel within which the counterparty is able to call updateTransfer and show the transfers they received.
- `merkletree root`:The root of the merkle tree which holds the hashes of all the locks in the channel.
- `locksroot`: The root of the merkle tree which holds the hashes of all the locks in the channel.
- `transferred amount`: The transferred amount is the total amount of token one participant of a channel has sent to his counterparty.
- `balance proof`: Balance proof is any kind of message we use in order to cryptographically prove to our counterparty (or them to us) that their balance has changed and that we have received a transfer.
- `DirectTransfer`: A direct transfer is a non-locked transfer, which means a transfer that does not rely on a lock to complete. Once they are sent they should be considered as completed.
- `MediatedTransfer`: A mediated transfer is a hashlocked transfer between an initiator and a target propagated through nodes in the network.
- `hashlock`: A hashlock is the hashed secret that accompanies a locked message: `sha3(secret)`
- `lock expiration`: The lock expiration is the highest block_number until which the transfer can be settled.
- `initiator`: In a mediated transfer the initiator of the transfer is the photon node which starts the transfer
- `target`: In a mediated transfer the target is the photon node for which the transfer sent by the initiator is intended
- `SecretRequest`: The secret request message is sent by the target of a mediated transfer to its initiator in order to request the secret to unlock the transfer.
- `RevealSecret`: The reveal secret message is sent to a node that is known to have an interest to learn the secret.
- `unlock message`: The unlock message is a message used for synchronization between mediated transfer participants.
- `secret`: The preimage, what we call the secret in photon, is 32 bytes  keccak hash ends up being the hashlock.
- `reveal timeout`: The number of blocks in a channel allowed for learning about a secret being reveal through the blockchain and acting on it.
- `SecretRequest`: The secret request message is sent by the target of a mediated transfer to its initiator in order to request the secret to unlock the transfer.
