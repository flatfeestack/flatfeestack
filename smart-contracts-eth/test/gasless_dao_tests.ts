import { expect } from "chai";
import { ethers } from "hardhat";
import { buildUserOp, PackedUserOperation } from "../scripts/buildUserOp";

describe("ERC-4337 Paymaster Local Test (TS)", function () {
    let entryPoint: any;
    let contractNFT: any;
    let contractDAO: any;
    let contractPaymaster: any;

    let council1: any;
    let council2: any;
    let user1: any;
    let user2: any;
    let relayer: any;

    before(async () => {
        [council1, council2, user1, user2, relayer] = await ethers.getSigners();

        const NFT = await ethers.getContractFactory("FlatFeeStackNFT");
        contractNFT = await NFT.deploy(
            council1.address,
            council1.address,
            council2.address
        );
        await contractNFT.waitForDeployment();

        const DAO = await ethers.getContractFactory("FlatFeeStackDAO");
        contractDAO = await DAO.deploy(await contractNFT.getAddress());
        await contractDAO.waitForDeployment();

        const EntryPoint = await ethers.getContractFactory(
            "@account-abstraction/contracts/core/EntryPoint.sol:EntryPoint"
        );
        entryPoint = await EntryPoint.deploy();
        await entryPoint.waitForDeployment();

        const Paymaster = await ethers.getContractFactory("FlatFeeStackDAOPaymaster");
        contractPaymaster = await Paymaster.deploy(
            await entryPoint.getAddress(),
            await contractNFT.getAddress(),
            await contractDAO.getAddress()
        );
        await contractPaymaster.waitForDeployment();

        // fund paymaster deposit to avoid validation failure
        await contractPaymaster.deposit({ value: ethers.parseEther("1") });
    });

    it("should allow a council member to perform a gasless userOp", async () => {
        const isCouncil = await contractNFT.isCouncil(1);
        expect(isCouncil).to.equal(true);

        const rawNonce = await entryPoint.getNonce(council1.address, 0);
        const nonce = BigInt(rawNonce);

        const callData = "0x";

        let userOp: PackedUserOperation = await buildUserOp(
            council1.address,
            nonce,
            callData
        );

        userOp.paymasterAndData = ethers.concat([
            await contractPaymaster.getAddress(),
            "0x00"
        ]);

        // Simulate gasless execution
        const tx = await entryPoint
            .connect(user2)
            .handleOps([userOp], user2.address);

        const receipt = await tx.wait();

        expect(receipt?.status).to.equal(1);
        console.log("Gasless userOp executed");
    });
});
