import {ethers, upgrades} from "hardhat";
import {deployMembershipContract} from "./helpers/deployContracts";
import {mine, time} from "@nomicfoundation/hardhat-network-helpers";
import {keccak256} from "@ethersproject/keccak256";
import {toUtf8Bytes} from "@ethersproject/strings";
import {expect} from "chai";


describe("DAO", () => {

    async function deployFixture() {
        const [owner, firstCouncilMember, secondCouncilMember, regularMember] = await ethers.getSigners();
        return await deployMembershipContract();
    }

    describe("propose", () => {
        it("cannot create a proposal if they don't have any votes", async () => {
            const fixtures = await deployFixture();
            const { dao, sbt} = fixtures.contracts;
            const { nonMember } = fixtures.entities;

            const transferCalldata = sbt.interface.encodeFunctionData(
                "pause",
                []
            );

            await expect(
                dao
                    .connect(nonMember)
                    .propose(
                        [sbt.address],
                        [0],
                        [transferCalldata],
                        "I would like to have some money to expand my island in Animal crossing."
                    )
            ).to.revertedWith("Governor: proposer votes below proposal threshold");
        });

        it("can propose a proposal", async () => {
            const fixtures = await deployFixture();
            const { dao, sbt} = fixtures.contracts;
            const { secondCouncilMember } = fixtures.entities;

            const transferCalldata = sbt.interface.encodeFunctionData(
                "pause", []
            );

            await expect(
                dao
                    .connect(secondCouncilMember)
                    .propose(
                        [sbt.address],
                        [0],
                        [transferCalldata],
                        "I would like to have some money to expand my island in Animal crossing."
                    )
            )
                .to.emit(dao, "ProposalCreated")
                .and.to.emit(dao, "DAOProposalCreated");
        });

    });

});