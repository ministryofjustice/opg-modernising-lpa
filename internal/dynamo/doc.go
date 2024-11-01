/*
Package dynamo provides a client for AWS DyanmoDB.

We use a combination of records for a full LPA. An example of the expected
set for an LPA is:

	| PK           | SK                             | Description                                                     | Type                             |
	| ------------ | ------------------------------ | --------------------------------------------------------------- | -------------------------------- |
	| LPA#...      | SUB#...                        | Links a OneLogin subject to all LPAs they can access            | dashboarddata.LpaLink            |
	| LPA#...      | DONOR#...                      | Data entered online by the donor                                | donordata.Provided               |
	| LPA#...      | RESERVED#DONOR#                | Ensure an LPA only has one donor                                |                                  |
	| LPA#...      | CERTIFICATE_PROVIDER#...       | Data entered online by the certificate provider                 | certificateproviderdata.Provided |
	| LPA#...      | RESERVED#CERTIFICATE_PROVIDER# | Ensure an LPA only has one certificate provider                 |                                  |
	| LPA#...      | ATTORNEY#...                   | Data entered online by an attorney (or replacement/trust corp.) | attorneydata.Provided            |
	| LPA#...      | VOUCHER#...                    | Data entered online by a voucher                                | voucherdata.Provided             |
	| LPA#...      | DOCUMENT#...                   | A document uploaded as evidence for a reduced fee               | document.Document                |
	| LPA#...      | EVIDENCE_RECEIVED#             | Marker to show paper evidence has been sent in to the OPG       |                                  |
	| UID#...      | METADATA#                      | Ensure a UID is only set once                                   |                                  |

For supporters there is data for the organisation, but also the LPA is stored against the donor differently:

	| PK               | SK               | Description                                                        | Type                       |
	| ---------------- | ---------------- | ------------------------------------------------------------------ | -------------------------- |
	| ORGANISATION#... | ORGANISATION#... | Holds data about the organisation                                  | supporterdata.Organisation |
	| ORGANISATION#... | MEMBER#...       | A member of an organisation                                        | supporterdata.Member       |
	| ORGANISATION#... | MEMBERINVITE#... | An invitation for a member to join an organisation                 | supporterdata.MemberInvite |
	| ORGANISATION#... | MEMBERID#...     | Allows querying a member by their ID, rather than OneLogin subject | supporter.organisationLink |
	| LPA#...          | ORGANISATION#... | Data entered for an LPA by a supporter at the organisation         | donordata.Provided         |
	| LPA#...          | DONOR#...        | Reference to the data accessible by the donor                      | donor.lpaReference         |

For sharing an LPA with each actor we generate records like:

	| PK                           | SK                       | Description                                                    | Type               |
	| ---------------------------- | ------------------------ | -------------------------------------------------------------- | ------------------ |
	| VOUCHERSHAREKEY#...          | VOUCHERSHARESORT#...     | A share of the LPA to a voucher                                | sharecodedata.Link |
	| DONORSHAREKEY#...            | DONORINVITE#...          | A share of an organisation created LPA to a donor              | sharecodedata.Link |
	| CERTIFICATEPROVIDERSHARE#... | METADATA#...             | A share of the LPA to a certificate provider                   | sharecodedata.Link |
	| ATTORNEYSHARE#...            | METADATA#...             | A share of the LPA to an attorney (or replacement/trust corp.) | sharecodedata.Link |

	 The scheduler uses the following structure:

	| PK               | SK            | Description                          | Type            |
	| ---------------- | ------------- | ------------------------------------ | --------------- |
	| SCHEDULEDDAY#... | SCHEDULED#... | An event to run on the specified day | scheduled.Event |
*/
package dynamo
