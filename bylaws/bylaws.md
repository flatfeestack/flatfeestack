# FlatFeeStack Association Bylaws
The bylaws as required by the Swiss law for the FlatFeeStack DAO.

## Name, Domicil, and Purpose

### Art. 1 Name
An association according to Art. 60 ff. ZGB exists with the name of "FlatFeeStack DAO".

### Art. 2 Domicile
FlatFeeStack DAO has its domicile in Zürich.

### Art. 3 Purpose
The purpose of FlatFeeStack DAO is to directly or indirectly contribute to the success of FlatFeeStack (platform) with the goal to support 
and promote open source software and its projects. Such activities can be of non-commercial as well as commercial nature. 
FlatFeeStack DAO is a not-for-profit association. FlatFeeStack DAO can cooperate with or join other organisations that represent 
the same or similar interests.

## Structure and Organization

### Art. 4 Bodies
FlatFeeStack DAO's bodies are:

- Council Member ("Vorstand");
- Ballot Vote ("Urabstimmung");

### Art. 5 General Concept of Competences and Duties

The objective of FlatFeeStack DAO is to establish a decentralized and democratic association with flat hierarchies.
Therefore, the Council Member ("Vorstand") will have only competences that require the action of an individual, natural person, 
such as representation duties or the duty to keep the books.

The Ballot Vote ("Urabstimmung") has elementary competences that are mandatory stipulated for an association assembly 
by Swiss law (such as the change of bylaws, the liquidation of the association, and others). The Ballot Vote will be held digital.
Additionally, the Ballot Vote shall be the body, which can decide about the "daily business" such as the proposal and 
support of projects and the funding allocation to projects. The Ballot Vote provides a blockchain-based technical infrastructure 
to efficiently, democratically and transparently propose and vote.

### Art. 6 Underlying Technology

FlatFeeStack DAO is technically built with Ethereum Smart Contracts. All voting will take place on this technical infrastructure.
The relevant technical functions are hereinafter written in italic. To be able to vote, holding a certain amount of Ether is 
necessary for every member (transaction fees, gas fees).

### Art. 7 Council Member (Vorstand)

FlatFeeStack DAO has at minimum 3 council members with the following competencies and duties:

1. Representing FlatFeeStack DAO to the outside world;
1. Keeping the books and creating the necessary financial statements of FlatFeeStack DAO;
1. Canceling a Ballot Vote with at least 2 Council Member signatures, if they have a good reason for it (`cancelVotingSlot`);
1. Executing contract calls on official orders by governmental authorities with at least 2 Council Member signatures.
1. Reviewing member applicants on their eligibility to join FlatFeeStack DAO.
   If the member meets the requirements of Art. 11, the Council Member will add their addresses to the member registry (`approveMembership`);
1. Remove Members who haven't paid their membership fees (`removeMembersThatDidntPay`);
1. Keeping the member registry (name, address, e-mail);

The term of office for the Council Member starts with the foundation of this FlatFeeStack DAO and ends when the Council Member has been 
replaced via a proposal. If a Council Member becomes unable to act or loses his private key, the Council Member or its address must 
be replaced with a new Council Member in the next Ballot Vote. If all the Council Members are unable to act, an extraordinary Ballot Vote 
will be called and two new Council Members must be elected. The personal liability of a Council Member is limited to cases of 
gross negligence.

### Art. 8 Ballot Vote ("Urabstimmung")

1. Competences: 
   The Ballot Vote shall be the highest governing body of FlatFeeStack DAO.
   The objective of the Ballot Vote is to allow every member to propose new projects and to vote on the funding allocation to those projects.
   The Ballot Vote has the duty of collecting all the votes on the proposals.

1. Voting Process:
   The whole voting process is purely digital and blockchain based.
   The members have one day to vote for every proposal in the Ballot Vote.

1. Voting Majorities:
   Resolutions shall be adopted by a simple majority of the members participating at the individual Ballot Vote.
   At least 1/5 of the Members must vote on a proposal and a simple majority to become successful.

1. Election Process for Council Member:
   Any member can propose themselves for candidacy as a council member (`propose` & `addCouncilMember`).
   If the proposal is successful the new Council Member term starts at the execution of the proposal.
   Any member can also propose to remove a council member (`propose` & `removeCouncilMember`).

1. Convocation:
   The Ballot Vote can be held in regular slots. The duration of a slot is 14 days and after 14 days, the slot 
   finishes and a new slot opens. There is no Ballot Vote in a slot if no proposal is submitted.
   The Council Member will inform the members electronically on the date, which is 7 days prior to the slot end.
   A Ballot Vote lasts one full day. Since every member can propose a Ballot Vote, no extraordinary Ballot Vote is necessary. 
   After a successful vote, the execution get delayed until the end of the slot.

1. Proposals & Agenda Items:
   Every member can submit a proposal. The contract execution via the Ballot Vote is binding and technically 
   non-reversible – not even by the Council Member. However, at least 2 Council Member signatures can cancel 
   a proposal before its execution, if they have a good reason. Proposals from members must be submitted at least 
   5 days prior to the Ballot Vote starting slot using the specific proposal function (`propose`).

## Membership

### Art. 9 Members & Membership Requirements

Natural persons, legal entities and organizations under public law can request membership to the FlatFeeStack DAO.
Legal entities and organizations under public law shall appoint a representative who exercises membership rights at the Ballot Vote.
Every member is responsible to gain the technological know-how to be able to participate on votes on the Ballot Vote.
In addition, every member has to assure to have a sufficient amount of Ether for the necessary transaction fees.

### Art. 10 Becoming a Member

Everyone who is eligible for a membership can make a request (`requestMembership`). Every member has to do a KYC check.
A new member has to be whitelisted by at least two Council Members before joining (`whitelistMember`).
After the whitelisting and the payment of the membership fee (`payMembershipFee`), the applicant becomes a Member 
and gains voting power for the next Ballot Vote.

### Art. 11 Awareness of Technological & Conceptual Risks
Blockchain is a new technology. The technical or conceptual structure of this FlatFeeStack DAO and the voting process may 
have weaknesses, as it is the case with every blockchain project. Moreover, the FlatFeeStack DAO is dependent on the 
underlying Ethereum protocol. Therefore, it may be possible that the FlatFeeStack DAO loses part or the whole of its 
funds or become incapable of acting. Every member explicitly declares to be aware of and to agree to those risks.

## Termination OF Membership

### Art. 12 Resignation
Members can leave FlatFeeStack DAO using `removeMember`. The resignation will be of immediate effect.
There is no entitlement to any refund of paid membership fees. The membership fee remains owed in full for 
the current fiscal year.

### Art. 13 Expiration
Membership in FlatFeeStack DAO ends automatically:

- upon liquidation of FlatFeeStack DAO;
- by the death of the specific member.

The membership fee remains owed in full for the current fiscal year.

### Art. 14 Exclusion

Every member of FlatFeeStack DAO can be expelled by the Member Community. The exclusion of a member can be 
proposed by every member (`propose` & `removeMember`). Members who did not pay their membership fees 
are excluded without voting process (`removeMembersThatDidntPay`). The excluded member has no right to an explanation.
The membership fee remains owed in full for the current fiscal year.

## Finances

### Art. 15 Membership Contributions and Other Fundraising
The FlatFeeStack DAO is primarily financed by the contributions of its members (`payMembershipFee`).
The Membership fee is 1 Ether. The Member contributions will be set initially or can be changed via a 
proposal (`propose` & `setMembershipFee`). In addition, the FlatFeeStack DAO is financed by a fee from 
the contributions from FlatFeeStack (platform). The membership fees are due in intervals of 365 days after the 
individual date of accession.

### Art. 16 Fiscal Year

The fiscal year is identical to the calendar year.

### Art. 17 Liability
The assets of FlatFeeStack DAO shall be solely liable for the obligations of FlatFeeStack DAO.
Personal liability of the members beyond the regularly adopted contributions is excluded.

## Assets
### Art. 18 Ownership Smart Contracts
The FlatFeeStack DAO has the ownership of following Ethereum Smart Contracts on the Mainnet:

- Name: DAA.sol Address: 0x 
- Name: Membership.sol Address: 0x 
- Name: Wallet.sol Address: 0x 

### Art. 19 Ownership Software Code
The FlatFeeStack DAO has the ownership of all the code who is written and stored on [GitHub](https://github.com/flatfeestack). 
All code must be licensed under an open source license.

## Update & Dissolution
### Art. 20 Update of the Underlying Code
An update of the underlying smart contract code of the DAA can be adopted via the Ballot Vote.

### Art. 21 Dissolution & Liquidation
The dissolution of FlatFeeStack DAO can be adopted with a proposal (`propose`).
The funds will be transferred to one of the Council Members (specified in the proposal), who must carry out the liquidation process.
If no Council Members are active, the funds can be transferred to a Member.

## Final Provisions
### Art. 22 Entry into Force
These Articles are based on the initial Ballot Vote of DD.MM.YYYY. They enter into force immediately.
