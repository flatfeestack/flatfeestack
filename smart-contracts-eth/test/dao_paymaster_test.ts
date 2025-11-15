import { Wallet } from "ethers";
const { expect } = require("chai");
const { ethers } = require("hardhat");

async function buildPaymasterAndData(
    paymaster: any,
    paymasterVerificationGas: any,
    paymasterPostOpGas: any
) {
    // Must ABI-encode the two uint128 values as the paymaster expects
    const encoded = ethers.AbiCoder.defaultAbiCoder().encode(
        ["uint128", "uint128"],
        [paymasterVerificationGas, paymasterPostOpGas]
    );

    // paymasterAndData = 20-byte paymaster address + encoded bytes
    return (await paymaster.getAddress()) + encoded.slice(2);
}

async function rawSignHash(signer: any, hash: any) {
  return await ethers.provider.send("eth_sign", [
    await signer.getAddress(),
    hash
  ]);
}

function buildUserOp(accountAddr: string, paymasterAndData: string) {

    const verificationGasLimit = 1_000_000n;
    const callGasLimit = 1_000_000n;

    // PACKED accountGasLimits = uint128 verification, uint128 call
    const accountGasLimits = ethers.solidityPacked(
        ["uint128", "uint128"],
        [verificationGasLimit, callGasLimit]
    );

    const maxFeePerGas = 10n;
    const maxPriorityFeePerGas = 10n;

    // PACKED gasFees = uint128 priorityFee, uint128 maxFee
    const gasFees = ethers.solidityPacked(
        ["uint128", "uint128"],
        [maxPriorityFeePerGas, maxFeePerGas]
    );

    const userOp = {
        // unpacked
        sender: accountAddr,
        nonce: 0n,
        initCode: "0x",
        callData: "0x",
        callGasLimit,
        verificationGasLimit,
        preVerificationGas: 50_000n,
        maxFeePerGas,
        maxPriorityFeePerGas,

        // packed (for EntryPoint)
        accountGasLimits,
        gasFees,

        paymasterAndData,
        signature: "0x"
    };

    return userOp;
}

describe("FlatFeeStack ERC4337 Integration", function () {

    let entryPoint : any;
    let nft : any;
    let dao : any;
    let paymaster : any;
    let factory : any;
    let owner : any, council1 : any, council2 : any, member : any;

    before(async function () {
        [owner, council1, council2, member] = await ethers.getSigners();

        // Deploy EntryPoint
        const EntryPoint = await ethers.getContractFactory("EntryPointSimulations");
        entryPoint = await EntryPoint.deploy();

        // Deploy DAO + NFT via Paymaster constructor
        const Paymaster = await ethers.getContractFactory("FlatFeeStackDAOPaymaster");
        const tx = await Paymaster.deploy(await entryPoint.getAddress(), council1.address, council2.address);

        const receipt = await tx.deploymentTransaction().wait();
        paymaster = await tx.waitForDeployment();

        const nftIface = new ethers.Interface([
            "event FlatFeeStackNFTCreated(address indexed addr, address indexed creator)"
        ]);

        let nftAddress = null;

        for (let log of receipt.logs) {
            try {
                const parsed = nftIface.parseLog(log);
                nftAddress = parsed.args.addr;
            } catch {}
        }

        expect(nftAddress).to.properAddress;

        const daoSigner = await ethers.getImpersonatedSigner(paymaster.target);
        nft = await ethers.getContractAt("FlatFeeStackNFT", nftAddress, daoSigner);

        // Top up paymaster so it can pay gas
        await owner.sendTransaction({
            to: await paymaster.getAddress(),
            value: ethers.parseEther("10")
        });

        // 3. Deploy SimpleAccount + Factory for ERC4337
        const Factory = await ethers.getContractFactory("SimpleAccountFactory");
        factory = await Factory.deploy(await entryPoint.getAddress());
    });

    it("Should mint member NFT and pay membership", async function () {
        const membershipFee = await nft.membershipFee();

        // next tokenId = 3
        const nextTokenId = (await nft.currentTokenId()) + 1n;

        // reproduce on-chain hashing logic
        const payloadHash = ethers.keccak256(
            ethers.solidityPacked(
                ["address", "string", "address", "string", "uint256"],
                [await nft.getAddress(), "safeMint", member.address, "#", nextTokenId]
            )
        );

        // sign using council wallets
        const sig1 = await council1.signMessage(ethers.getBytes(payloadHash));
        const sig2 = await council2.signMessage(ethers.getBytes(payloadHash));

        // safeMint with council signatures
        await nft.connect(member).safeMint(
            member.address,
            0, sig1,
            0, sig2,
            { value: membershipFee }
        );

        expect(await nft.balanceOf(member.address)).to.equal(1n);
    });

    it("Should create a SimpleAccount for member", async function () {
        let tx = await factory.createAccount(member.address, 0);
        let receipt = await tx.wait();
        let accountAddr = await factory.getAddress(member.address, 0);

        expect(accountAddr).to.properAddress;
    });

    /*it("Should validate Paymaster UserOp", async function () {
        let accountAddr = await factory.getAddress(member.address, 0);

        const paymasterVerificationGas = 50_000n;
        const paymasterPostOpGas = 50_000n;

        const paymasterAndData = ethers.hexlify(
            ethers.concat([
                ethers.getBytes(await paymaster.getAddress()),
                ethers.zeroPadValue(ethers.toBeHex(paymasterVerificationGas), 16),
                ethers.zeroPadValue(ethers.toBeHex(paymasterPostOpGas), 16)
            ])
        );

        const userOp = buildUserOp(accountAddr, paymasterAndData);
        const userOpHash = await entryPoint.getUserOpHash(userOp);
        userOp.signature = await rawSignHash(member, userOpHash);

        const recovered = ethers.recoverAddress(userOpHash, userOp.signature);
        console.log("RECOVERED:", recovered);
        console.log("EXPECTED :", await member.getAddress());

        console.log("DEBUG: userOp", userOp);
        console.log("DEBUG: sender code size",
            (await ethers.provider.getCode(userOp.sender)).length
        );
        
        // Call validation
        try {
            await entryPoint.simulateValidation.staticCall(userOp);
        } catch (err) {
            console.log("SIMULATION ERROR RAW:", err);
            throw err;
        }
    });

    it("Should allow gasless execution through EntryPoint", async function () {
        let accountAddr = await factory.getAddress(member.address, 0);

        // Real userOp encoding omitted for clarity → in full tests you use UserOp builder
        let userOp = {
            sender: accountAddr,
            nonce: 0,
            initCode: "0x",
            callData: "0x",
            callGasLimit: 1_000_000n,
            verificationGasLimit: 1_000_000n,
            preVerificationGas: 100_000n,
            maxFeePerGas: 10n,
            maxPriorityFeePerGas: 10n,
            paymasterAndData: await paymaster.getAddress(),
            signature: "0x"
        };

        await expect(
            entryPoint.handleOps([userOp], owner.address)
        ).to.not.be.reverted;
    });*/
});
