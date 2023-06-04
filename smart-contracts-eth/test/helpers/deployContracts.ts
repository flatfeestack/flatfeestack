import { ethers, upgrades } from "hardhat";
import { mine } from "@nomicfoundation/hardhat-network-helpers";

export async function deployMembershipContract(){

  const [owner, firstCouncilMember, secondCouncilMember,regularMember, nonMember ] = await ethers.getSigners();

  const SBT = await ethers.getContractFactory("FlatFeeStackDAOSBT", {signer: owner});
  const sbt = await upgrades.deployProxy(SBT);
  await sbt.deployed();

  await sbt.connect(owner).safeMintCouncil(firstCouncilMember.address, 1);
  await sbt.connect(owner).safeMintCouncil(regularMember.address, 2);
  //here we test a replacement
  await sbt.connect(owner).safeMintCouncil(secondCouncilMember.address, 2);

  const DAO = await ethers.getContractFactory("FlatFeeStackDAO", {signer: owner});
  const dao = await upgrades.deployProxy(DAO, [sbt.address]);
  await dao.deployed();

  await sbt.connect(owner).grantRole(ethers.utils.formatBytes32String("0x00"),dao.address);
  await sbt.connect(owner).revokeRole(ethers.utils.formatBytes32String("0x00"),owner.address);

  //add the regular member now
  const message=ethers.utils.solidityPack(['address', 'string', 'address', 'uint256'],[sbt.address, 'safeMint', regularMember.address, 100])
  const hash = ethers.utils.solidityKeccak256(["bytes"], [message]);

  const signedMessage1 = await firstCouncilMember.signMessage(ethers.utils.arrayify(hash));
  const signedMessage2 = await secondCouncilMember.signMessage(ethers.utils.arrayify(hash));

  await sbt.connect(regularMember).safeMint(
      regularMember.address,
      100,
      signedMessage1,
      signedMessage2,
      { value: ethers.utils.parseEther('1') });

  //see https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/governance/utils/Votes.sol
  await sbt.connect(regularMember).delegate(regularMember.address);
  await sbt.connect(firstCouncilMember).delegate(firstCouncilMember.address);
  await sbt.connect(secondCouncilMember).delegate(secondCouncilMember.address);

  await mine(2);

  return {
    contracts: {
      dao,
      sbt
    }, entities: {
      owner,
      firstCouncilMember,
      secondCouncilMember,
      regularMember,
      nonMember
    }
  }
}
