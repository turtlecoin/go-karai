# Karai Transactions

Transactions in Karai channels are simple JSON objects with a few key shared fields. Different transactions types will have different roles and functions. This guide will explain the different transaction types and the functions of their roles.

**The 4 transaction types:**

-   **(Type 0)** - Root Tx
-   **(Type 1)** - Milestone Tx
-   **(Type 2)** - Subgraph Tip Tx
-   **(Type 3)** - Normal Tx

## Shared Characteristics of Transactions

Transactions regardless of type will contain a few shared common elements:

##### Cryptographic Elements

-   A pubkey
-   A signature
-   A signed msg

##### Transaction Elements

-   A transaction type
-   A transaction hash

# Root Tx Characteristics

The purpose of the Root Tx is initiating a transaction channel and beginning the merkle hash tree.

##### Quick facts:

-   The Root Tx is transaction type 0.
-   The Root Tx is always the first transaction in a Karai transaction channel.
-   Is immediately followed by a Milestone transaction always.
-   The Root Tx is the smallest transaction possible in a transaction channel.
-   The Root Tx is the only transaction type with no `tx_prev` field.

```json
[
    {
        "coord_pub": "cfae4ecca8ed282ab51ae16bf755e510b1ce52d6e84263f357b0a524691e9259",
        "coord_sig": "2efc5209e8f19c7d0f50ac08107b0c5690ba173c6371d8d9a9b46b1790b4e709ce7e492671cb76203ff4b40e40bbaf2075db7741eeae8b4a641fe36bfa8b880b",
        "coord_msg": "48a49aa318fa36ee623bae1b25e23d87a21923d889f620e62106009699de17fc"
    },
    {
        "tx_type": 0,
        "tx_hash": "1ae3529687d8340213f6ffc2c5a4d2747ff35575eb12841e7bfea60660cd69c1",
        "tx_data": "Karai Transaction Channel - Root",
        "tx_time": 1588544439332036619
    }
]
```

#### `coord_pub`

Channel coordinator public TRTL key.

#### `coord_sig`

Channel coordinator signature.

#### `coord_msg`

Channel coordinator signed message.

#### `tx_type`

Defines this transaction's type. This transaction will always be type 0

#### `tx_hash`

This is a SHA256 hash of the root transaction's contents.

#### `tx_data`

This is an arbitrary data field.

#### `tx_time`

UNIX time in nanoseconds. This begins the lineage of the merkle hash and along with the tx_data field adds another factor of randomness to the hashing process of the Root Tx.

# Milestone Tx Characteristics

The purpose of the Milestone Tx is initiating and modifying the governance policy of the channel. Any changes to the way a channel performs subgraph construction or triages transactions will happen via a Milestone Tx.

##### Quick facts:

-   The Milestone Tx is transaction type 1.
-   The Milestone Tx is always the second transaction in a Karai transaction channel.
-   Beyond the first Milestone Tx following the initial creation of the channel, Milestone channels are not required or generated automatically.
-   A Milestone Tx is always referenced in a Subgraph Tip Tx or a Normal Tx to define the ruleset governing their behavior.

```json
[
    {
        "coord_pub": "cfae4ecca8ed282ab51ae16bf755e510b1ce52d6e84263f357b0a524691e9259",
        "coord_sig": "2efc5209e8f19c7d0f50ac08107b0c5690ba173c6371d8d9a9b46b1790b4e709ce7e492671cb76203ff4b40e40bbaf2075db7741eeae8b4a641fe36bfa8b880b",
        "coord_msg": "48a49aa318fa36ee623bae1b25e23d87a21923d889f620e62106009699de17fc"
    },
    {
        "milestone_id": 0,
        "milestone_hash": "ecbe9ed98d8fc513cc6e508ee20ac06dfe7178ad3855e2afbd626dfe7178ad3a",
        "milestone_prev": "",
        "wavetip_matrix": [
            {
                "index": 0,
                "tx": "dfe7178ad3855e2afbd626a7ca38975b901591859c2cdbe6da559dfe8ef3bc4a"
            },
            {
                "index": 1,
                "tx": "450d303bfb69d8b58a362cd539e951cb88fa886bff6d0218f99db23f2d0a2c91"
            }
        ],
        "participant_matrix": [
            {
                "index": 0,
                "participants": [
                    "77a4781055c9d26c3136afefa7823037e898f92105bcaaeac385719d42b50f20",
                    "b7e0d0839883bf207b39e032343fa1569dbfeb6075c9fa9e5de41e025c367e90",
                    "6eab8b2f2e55e0b10a527e2a9ec85acc3c89e3f9cb92fa30e9af0bc49b5181c3"
                ]
            },
            {
                "index": 1,
                "participants": [
                    "7e43c35159d57c2a72c9b6c2fecb03b2df037df7d71e1817bcbad8983cf118c0",
                    "fd8c4686d6d82b78cc0a3bd09ff165c97a6eb1270279d862b0d544eda6b8fc68",
                    "588e9d3ef6ef2725ed8795c78b45675f892d88cba27856665150a17da974bc63"
                ]
            }
        ]
    },
    {
        "tx_type": 1,
        "tx_hash": "a0c1d8ccae6271a1906c4a7b881ce8d0808702ad506d400e8cb1a14f15063c08",
        "tx_prev": "1ae3529687d8340213f6ffc2c5a4d2747ff35575eb12841e7bfea60660cd69c1",
        "tx_data": {
            "chan_milestone": 0,
            "channel_params": {
                "name": "karai-transaction-channel",
                "public": true,
                "n_txspread": 1,
                "n_interval": 30
            },
            "coord_params": {
                "majorSemver": 0,
                "minorSemver": 2,
                "patchSemver": 1,
                "address": "TRTLuxMpUNTBqfWshbc65E7Yqx17rQpHQZC5HyANaTL37AQm2fRsNDXG37jxPhXXa5NMJVLFJpQa9iQn9Se87VNuWwPHWScoZLY"
            },
            "peer_data": {
                "ip4": "123.34.45.56",
                "ifps_peer_id": "QmfMAVSMyw4T9Mu3y8hM4phLWbM5NhdYuq5HRmKy8kX3SD",
                "ipfs_pub_key": "CAASpgIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDI3gvSHQ/V3o7wWLp+KLw8w4k74JGF7+lxPAK0Z6SAp2CELvr+FJfflcIAnOna5NekFj3oZhgI3sTAMRixn802S+OUmBuFrtdxd8SX1PjwCmdzm+xTWU8IdrZbxzeHY/n4i34ZyOEybdWEvR4oExplxTk9mnZmKvZvIH3lCQIbfkhoJFTB4D4R5KG5YcEQ6/2hLvzdoMyUcVZRf7dRxWUyoXRdE5810tsCBECrRzLX9nWERP/ki4elvJlDQYU5bHUazZy4tbl9kEbP28gjm9XGYKxjAWyXG+uMZoohCujSNN3SQzo/5zE4VWzi4LC01ourl8xR9pd5HhzH1oKcYBoZAgMBAAE="
            }
        }
    }
]
```

#### `coord_pub`

Channel coordinator public TRTL key.

#### `coord_sig`

Channel coordinator signature.

#### `coord_msg`

Channel coordinator signed message.

#### `milestone_id`

This milestone ID.

#### `milestone_hash`

This milestone hash.

#### `milestone_prev`

The previous milestone hash.

#### `wavetip_matrix`

A matrix of the constituent subgraph tips from this completed milestone (if this milestone is complete).
The matrix consists of two data points:

-   index
-   tx

#### `participant_matrix`

A matrix of the participant ID data for this milestone. For each `index` corresponding with a point in the `wavetip_matrix`, list every participant in the subgraph.

#### `tx_type`

Defines this transaction's type. This transaction will always be type 1.

#### `tx_hash`

This is a SHA256 hash of this milestone transaction's contents.

#### `tx_prev`

This is the previous hash

#### `tx_data`

This is an arbitrary data field, in this case it holds the ruleset for subsequent Subgraphs:

#### `tx_data` > `channel_milestone`

0-indexed count of all milestones including this milestone since the creation of the channel.

#### `tx_data` > `channel_params`

These are the governance changes being passed:

-   `name` - The name of the channel. This is used for identifying a channel by human readable name in the Karai client.
-   `public` - This is set to true if a channel has published a pointer to a blockchain and wishes to be publicly listed. This value does not have any auth related implications.
-   `n_txspread` - In "graph mode", a transaction in a subgraph will have a capacity of `1 + n_txspread` number of child transactions. A value of 0 here would produce a linear chain.
-   `n_interval` - This is a value measured in seconds. After `n_interval` seconds of the channel coordinator not receiving a transaction, the first transaction received will start a timer for `n_interval` seconds to 'listen' for other tx's that can be assembled as members of a subgraph. When this time has passed, all of the transactions received during the interval are arranged according to `n_txspread` and recorded as a completed subgraph by the channel coordinator.

#### `tx_data` > `coord_params`

This is data recording the status of the channel coordinator:

-   `majorSemver` - Major version number.
-   `minorSemver` - Minor version number.
-   `patchSemver` - Patch version number.
-   `address` - TRTL receive address for payments to this coordinator.

#### `tx_data` > `peer_data`

This is peer data for the channel coordinator:

-   `ip4` - IP v4 address of the channel coordinator
-   `ipfs_peer_id` - IPFS peer ID
-   `ipfs_pub_key` - IPFS public key

# Subgraph Tip Tx Characteristics

The Subgraph Tip Tx is the chronologically first transaction in a completed subgraph. It can be the only transaction in the subgraph if the `n_interval` is low or if Tx volume is inconsistent.

##### Quick facts:

-   The Subgraph Tip is transaction type 2.
-   In the past, the Subgraph tip has been called the Wave Tip.
-   The Subgraph Tip Tx is the first Tx received in a subgraph collection interval.

```json
[
  {
      "coord_pub": "cfae4ecca8ed282ab51ae16bf755e510b1ce52d6e84263f357b0a524691e9259",
      "coord_sig": "2efc5209e8f19c7d0f50ac08107b0c5690ba173c6371d8d9a9b46b1790b4e709ce7e492671cb76203ff4b40e40bbaf2075db7741eeae8b4a641fe36bfa8b880b",
      "coord_msg": "48a49aa318fa36ee623bae1b25e23d87a21923d889f620e62106009699de17fc"
  },
  {
      "milestone_id": 0,
      "milestone_hash": "ecbe9ed98d8fc513cc6e508ee20ac06dfe7178ad3855e2afbd626dfe7178ad3a",
      "milestone_prev": "",
      "subgraph_id": 0,
      "tx_matrix": [
          {
              "index": 0,
              "tx": "dfe7178ad3855e2afbd626a7ca38975b901591859c2cdbe6da559dfe8ef3bc4a"
          },
          {
              "index": 1,
              "tx": "450d303bfb69d8b58a362cd539e951cb88fa886bff6d0218f99db23f2d0a2c91"
          }
      ],
      "participant_matrix": [
          {
              "index": 0,
              "participants": [
                  "77a4781055c9d26c3136afefa7823037e898f92105bcaaeac385719d42b50f20",
                  "b7e0d0839883bf207b39e032343fa1569dbfeb6075c9fa9e5de41e025c367e90",
                  "6eab8b2f2e55e0b10a527e2a9ec85acc3c89e3f9cb92fa30e9af0bc49b5181c3"
              ]
          },
          {
              "index": 1,
              "participants": [
                  "7e43c35159d57c2a72c9b6c2fecb03b2df037df7d71e1817bcbad8983cf118c0",
                  "fd8c4686d6d82b78cc0a3bd09ff165c97a6eb1270279d862b0d544eda6b8fc68",
                  "588e9d3ef6ef2725ed8795c78b45675f892d88cba27856665150a17da974bc63"
              ]
          }
      ]
  },
  {
    "tx_type": 2,
    "tx_hash": "17b66b01d4957989429217f51453a5c6916fbda085e6d487c42e8c1542e6fcaa",
    "tx_prev": "a0c1d8ccae6271a1906c4a7b881ce8d0808702ad506d400e8cb1a14f15063c08",
    "tx_data": {}
  }
```

#### `coord_pub`

Channel coordinator public TRTL key.

#### `coord_sig`

Channel coordinator signature.

#### `coord_msg`

Channel coordinator signed message.

#### `milestone_id`

The ID of the current milestone governing the construction of this subgraph.

#### `milestone_hash`

The hash of the milestone governing the construction of this subgraph.

#### `milestone_prev`

The previous hash of the milestone governing the construction of this subgraph.

#### `subgraph_id`

The ID of this subgraph

#### `tx_matrix`

The matrix of transactions in this subgraph

#### `participant_matrix`

The matrix of participants in this subgraph

#### `tx_type`

Defines this transaction's type. This transaction will always be type 2.

#### `tx_hash`

This is a SHA256 hash of this transaction's contents.

#### `tx_data`

This is an arbitrary data field

# Standard Tx Characteristics

This is a standard transaction. It is the most basic transaction you can send.

```json
[
    {
        "coord_pub": "cfae4ecca8ed282ab51ae16bf755e510b1ce52d6e84263f357b0a524691e9259",
        "coord_sig": "2efc5209e8f19c7d0f50ac08107b0c5690ba173c6371d8d9a9b46b1790b4e709ce7e492671cb76203ff4b40e40bbaf2075db7741eeae8b4a641fe36bfa8b880b",
        "coord_msg": "48a49aa318fa36ee623bae1b25e23d87a21923d889f620e62106009699de17fc"
    },
    {
        "participant_pub": "cf755e510b1ce52d6e84263f357b0a524691e925fae4ecca8ed282ab51ae16b9",
        "participant_sig": "5690ba173c6371d8d9a9b46b1790b4e709ce7e492671cb76203ff4b40e40bbaf2075db7741eeae8b2efc5209e8f19c7d0f50ac08107b0c4a641fe36bfa8b880b",
        "participant_msg": "21923d889f620e6210600948a49aa318fa36ee623bae1b25e23d87a699de17fc"
    },
    {
        "tx_type": 3,
        "tx_hash": "a0c1d8ccae6271a1906c4a7b881ce8d0808702ad506d400e8cb1a14f15063c08",
        "tx_prev": "1ae3529687d8340213f6ffc2c5a4d2747ff35575eb12841e7bfea60660cd69c1",
        "tx_data": {}
    }
]
```

#### `coord_pub`

Channel coordinator public TRTL key.

#### `coord_sig`

Channel coordinator signature.

#### `coord_msg`

Channel coordinator signed message.

#### `participant_pub`

Channel participant public TRTL key.

#### `participant_sig`

Channel participant signature.

#### `participant_msg`

Channel participant signed message.

#### `tx_type`

Defines this transaction's type. This transaction will always be type 3.

#### `tx_hash`

This is a SHA256 hash of this transaction's contents.

#### `tx_data`

This is an arbitrary data field
