# FlatFeeStack Association Bylaws
The bylaws as required by the Swiss law for the FlatFeeStack DAO.
 
## Name, Domicil, and Purpose

### Art. 1 Name
An association according to Art. 60 ff. ZGB exists with the name of "FlatFeeStack DAO".

### Art. 2 Domicile
FlatFeeStack DAO has its domicile in Zürich.

### Art. 3 Purpose
The purpose of FlatFeeStack DAO is to directly or indirectly contribute to the success of FlatFeeStack (platform) with the goal to support and promote open source software and its projects. Such activities can be of non-commercial as well as commercial nature. FlatFeeStack DAO is a not-for-profit association. FlatFeeStack DAO can cooperate with or join other organisations that represent the same or similar interests.

## Structure and Organization

### Art. 4 Bodies
FlatFeeStack DAO's bodies are:

- Council Member ("Vorstand");
- Ballot Vote ("Urabstimmung");

### Art. 5 General Concept of Competences and Duties

The objective of FlatFeeStack DAO is to establish a decentralized and democratic association with flat hierarchies. Therefore, the Council Member ("Vorstand") will have only competences that require the action of an individual, natural person, such as representation duties, managing financial records, or complying with governmental orders.

The Ballot Vote ("Urabstimmung") has elementary competences that are mandatory stipulated for an association assembly by Swiss law (such as the change of bylaws, the liquidation of the association, and others). The Ballot Vote will be held digital. Additionally, the Ballot Vote shall be the body, which can decide about the "daily business" such as the proposal and support of projects and the funding allocation to projects. The Ballot Vote provides a blockchain-based technical infrastructure to transparently propose and vote.

### Art. 6 Underlying Technology

FlatFeeStack DAO is technically built with Solidity Smart Contracts. All voting will take place on this technical infrastructure. The relevant technical functions are hereinafter written in italic. To be able to vote, holding a certain amount of crypto currency is necessary for every member (transaction fees, gas fees).

### Art. 7 Council Member (Vorstand)

FlatFeeStack DAO has at minimum 2 council members with the following competencies and duties:

1. Representing FlatFeeStack DAO to the outside world.
1. Managing financial records and creating the necessary financial statements of FlatFeeStack DAO.
1. Canceling a Ballot Vote with exactly 2 Council Member signatures, if the council has good reason (`councilCancel`).
1. Executing contract calls on official orders by governmental authorities with exactly 2 Council Member signatures (`councilExecute`).
1. Reviewing member applicants on their eligibility to join FlatFeeStack DAO. If the member meets the requirements of Art. 11, the Council Member will endorse the member by providing two council signatures. This authorization enables the new member to generate their membership non-fungible token (NFT) (`safeMint`).
1. Remove Members who haven't paid their membership fees (`burn`).
1. Keeping the member registry (name, address, e-mail).

The term of office for the Council Member starts with the foundation of this FlatFeeStack DAO and ends when the Council Member has been replaced via a proposal.If a Council Member becomes unable to act or loses his private key, the Council Member or its address must be replaced with a new Council Member in the next Ballot Vote. If all the Council Members are unable to act, an new Ballot Vote will be called and two new Council Members must be elected. The personal liability of a Council Member is limited to cases of gross negligence.

### Art. 8 Ballot Vote ("Urabstimmung")

1. Competences: 
   The Ballot Vote shall be the highest governing body of FlatFeeStack DAO. The objective of the Ballot Vote is to allow every member to propose new projects and to vote on proposals, such as the funding allocation to those projects. The Ballot Vote automatically collects all the votes on the proposals.

1. Voting Process:
   The voting process is digital and blockchain based.

1. Voting Majorities:
   Proposals shall be adopted by a simple majority of the Members participating at the individual Ballot Vote. Vote participation is only considered for approving or abstaining votes. In case of 2 Members, a single vote is sufficient. For more than 2 Members, at least 2 votes are required to reach a majority, and the number of required votes is calculated as follows: min_votes = max(2, int(20% * total_members) / total_members), indicating a vote participating of ~14-20% (depending on rounding).

1. Election Process for Council Member:
   Any member can propose themselves for candidacy as a Council Member (`propose` & `setCouncil`). If the proposal is successful the new Council Member term starts at the execution of the proposal. Any member can also propose to remove a Council Member (`propose` & `setCouncil`).

1. Convocation:
   The Ballot Vote starts with a proposal submitted in regular slots. The duration of a slot is 7 days. After one slot finishes, the next immediately begins. The voting process starts after one full slot finishes. There is no Ballot Vote in a slot if no proposal is submitted. The Council Member will inform the Members electronically on the date, which is 3 days prior to the slot end. A Ballot Vote lasts 1 day. Since every member can propose a Ballot Vote, no extraordinary Ballot Vote is necessary. After a successful vote, the execution gets queued until the end of the next slot.

1. Proposals & Agenda Items:
   Every member can submit a proposal. The contract execution via the Ballot Vote is binding and technically non-reversible – not even by the Council Member. However, at least 2 Council Member signatures can cancel a proposal before its execution, if they have a good reason.

## Membership

### Art. 9 Members & Membership Requirements

Natural persons, legal entities and organizations under public law can request membership to the FlatFeeStack DAO. Legal entities and organizations under public law shall appoint a representative who exercises membership rights at the Ballot Vote. Every member is responsible to gain the technological know-how to be able to participate on votes on the Ballot Vote. In addition, every member has to assure to have a sufficient amount of Ether for the necessary transaction fees.

### Art. 10 Becoming a Member

To become a new member of FlatFeeStack DAO, an applicant must receive approval from at least 2 Council Members that will endorse the member by providing two council signatures. Once approved, the applicant can request a membership NFT (`safeMint`), formalizing their membership status. This process grants the new Member voting rights for the next Ballot Vote.

### Art. 11 Awareness of Technological & Conceptual Risks

Blockchain is a new technology. The technical or conceptual structure of this FlatFeeStack DAO and the voting process may have weaknesses, as it is the case with every blockchain project. Moreover, the FlatFeeStack DAO is dependent on the underlying Blockchain protocol. Therefore, it may be possible that the FlatFeeStack DAO loses part or the whole of its funds or become incapable of acting. Every member explicitly declares to be aware of and to agree to those risks.

## Termination OF Membership

### Art. 12 Resignation

Members can leave FlatFeeStack DAO using (`burn`). The resignation will be of immediate effect. There is no entitlement to any refund of paid membership fees. The membership fee remains owed in full for the current fiscal year.

### Art. 13 Expiration

Membership in FlatFeeStack DAO ends automatically:

- upon liquidation of FlatFeeStack DAO.
- by the death of the specific member.

The membership fee remains owed in full for the current fiscal year.

### Art. 14 Exclusion

Every member of FlatFeeStack DAO can be expelled by the Ballot Vote. The exclusion of a Member can be proposed by every member (`propose` & `burn`). Members who did not pay their membership fees can be excluded without voting process (`burn`). The excluded member has no right to an explanation. The membership fee remains owed in full for the current fiscal year.

## Finances

### Art. 15 Membership Contributions and Other Fundraising

The FlatFeeStack DAO is primarily financed by the contributions of its members (`safeMint`, `payMembership`). The Membership fee is 1 Ether. The Member contributions will be set initially or can be changed via a proposal (`propose` & `setMembershipSettings`). In addition, the FlatFeeStack DAO can be financially supported by a fee from the contributions from FlatFeeStack (platform). The membership fees are due in intervals of 365 days after each Member's individual start date.

### Art. 16 Fiscal Year

A fiscal year is 365 days, starting with the smart contract creation on the blockchain.

### Art. 17 Liability

The assets of FlatFeeStack DAO shall be solely liable for the obligations of FlatFeeStack DAO. Personal liability of the members beyond the regularly adopted contributions is excluded.

## Assets

### Art. 18 Ownership Smart Contracts
The FlatFeeStack DAO has the ownership of following Smart Contracts on the XYZ (mainnet):

- Name: FlatFeeStackDAO.sol Address: 0x 
- Name: FlatFeeStackNFT.sol Address: 0x 

### Art. 19 Ownership Software Code

The FlatFeeStack DAO has the ownership of all the code who is written and stored on [GitHub](https://github.com/flatfeestack). All source code must be licensed under an open source license.

## Update & Dissolution

### Art. 20 Update of the Underlying Code

The FlatFeeStack DAO's smart contract cannot be updated directly. However, deploying a new DAO smart contract allows existing NFT memberships to be counted under this new contract, with any transfer of remaining assets subject to a Ballot Vote. 

To amend the bylaws of the FlatFeeStack DAO, it's necessary to update the SHA256 hash in the DAO's smart contract. This process additionally mandates a Ballot Vote for approval (`setNewBylawsHash`).

### Art. 21 Dissolution & Liquidation

The dissolution of FlatFeeStack DAO can be adopted with a proposal (`propose`, `pause`). Any remaining funds need to be transferred to one of the Council Members (specified in the proposal), who must carry out the liquidation process. If no Council Members are active, the funds can be transferred to a Member.

## Final Provisions
### Art. 22 Entry into Force

These Articles are based on the initial contract creation of DD.MM.2024. They enter into force immediately.