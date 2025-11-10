import { expect } from "chai";
import { ethers } from "hardhat";
import type { Signer } from "ethers";

// replicated from Paymaster_test.sol
const ticketType = {
    Ticket: [
      { name: "sender",   type: "address" },
      { name: "target",   type: "address" },
      { name: "maxCost",  type: "uint256" },
      { name: "deadline", type: "uint48"  },
      { name: "chainId",  type: "uint256" },
    ],
  };

function buildCallDataToExecute(to: string, value: bigint, data: string): string {
  return ethers.AbiCoder.defaultAbiCoder().encode(
    ["address", "uint256", "bytes"],
    [to, value, data]
  );
}

function prefillUserOpWithDefaults(userOp: any): any {
  const defaults = {
    sender: ethers.ZeroAddress,
    nonce: 0,
    initCode: "0x",
    callData: "0x",
    callGasLimit: 1_000_000,
    verificationGasLimit: 1_000_000,
    preVerificationGas: 50_000,
    maxFeePerGas: ethers.parseUnits("1", "gwei"),
    maxPriorityFeePerGas: ethers.parseUnits("1", "gwei"),
    paymasterAndData: "0x",
    signature: "0x",
  };
  
  return { ...defaults, ...userOp };
}

async function signWithdrawMessage(payoutEth: any, owner: any,
    userId: string, totalPayoutWei: bigint): Promise<string> {
  const payloadHash = ethers.solidityPackedKeccak256(
    ["address", "string", "uint256", "string", "uint256"],
    [await payoutEth.getAddress(), "calculateWithdraw", userId, "#", totalPayoutWei]
  );

  const sig = await owner.signMessage(ethers.getBytes(payloadHash));

  const recovered = ethers.verifyMessage(ethers.getBytes(payloadHash), sig);
  if (recovered.toLowerCase() !== (await owner.getAddress()).toLowerCase()) {
    throw new Error("withdraw EIP-191 signature mismatch");
  }

  return sig;
}

async function buildPaymasterAndData(paymaster: any, pmAuthority: any,
  ticket: { sender: string; target: string; maxCost: bigint; deadline: number; chainId: bigint | number }
): Promise<string> {
  const domain = {
    name: "Paymaster",
    version: "1",
    chainId: ticket.chainId,
    verifyingContract: await paymaster.getAddress(),
  };

  const pmSig = await pmAuthority.signTypedData(domain, ticketType, ticket);

  // verify signer matches
  const recovered = ethers.verifyTypedData(domain, ticketType, ticket, pmSig);
  if (recovered.toLowerCase() !== (await pmAuthority.getAddress()).toLowerCase()) {
    throw new Error("paymaster EIP-712 signature mismatch");
  }

  const encoded = ethers.AbiCoder.defaultAbiCoder().encode(
    ["tuple(address,address,uint256,uint48,uint256)", "bytes"],
    [[ticket.sender, ticket.target, ticket.maxCost, ticket.deadline, ticket.chainId], pmSig]
  );

  return (await paymaster.getAddress()) + encoded.slice(2);
}

async function signUserOp(entryPoint: any, userOp: any, accountOwner: any): Promise<string> {
  const userOpHash = await entryPoint.getUserOpHash(userOp);
  return accountOwner.signMessage(ethers.getBytes(userOpHash));
}

describe("ERC-4337 Account Abstraction gasless", () => {
  let owner: Signer;
  let pmAuthority: Signer;
  let beneficiary: Signer;
  let user: Signer;

  let entryPoint: any;
  let account: any;
  let paymaster: any;
  let payoutEth: any;

  beforeEach(async () => {
    [owner, pmAuthority, beneficiary, user] = await ethers.getSigners();

    const EP = await ethers.getContractFactory("TestEntryPoint");
    entryPoint = await EP.deploy();
    await entryPoint.waitForDeployment();

    const PayoutEth = await ethers.getContractFactory("PayoutEth");
    payoutEth = await PayoutEth.deploy();
    await payoutEth.waitForDeployment();

    // fund payout with 10 ETH
    await (owner as any).sendTransaction({
      to: await payoutEth.getAddress(),
      value: ethers.parseEther("10")
    });

    const Account = await ethers.getContractFactory("FirstAccount");
    account = await Account.deploy(
      await owner.getAddress(),
      await entryPoint.getAddress()
    );
    await account.waitForDeployment();

    const PM = await ethers.getContractFactory("Paymaster");
    paymaster = await PM.deploy(
      await entryPoint.getAddress(),
      await pmAuthority.getAddress()
    );
    await paymaster.waitForDeployment();

    // to cover gas, fund paymaster
    await paymaster.connect(owner).deposit({ value: ethers.parseEther("5") });

    await paymaster
      .connect(owner)
      .setTargetAllowed(await payoutEth.getAddress(), true);
  });

  it("runs a gasless userOperation via EntryPoint.handleOps", async () => {
    const userAddress = await user.getAddress();
    const accountAddress = await account.getAddress();
    const payOutAddress = await payoutEth.getAddress();
    const paymasterAddress = await paymaster.getAddress();

    const payOutAmount = ethers.parseEther("1");
    const userId = ethers.sha256(ethers.getBytes(userAddress));
    const chainId = (await ethers.provider.getNetwork()).chainId;

    const withdrawSig = await signWithdrawMessage(payoutEth, owner, userId, payOutAmount);

    // combine final data to call the withdraw function
    const targetCalldata = payoutEth.interface.encodeFunctionData("withdraw", [
        userAddress,
        userId,
        payOutAmount,
        withdrawSig,
    ]);
    const callData = buildCallDataToExecute(payOutAddress, 0n, targetCalldata);

    // user operation with mostly defaults
    const partialUserOp = {
        sender: accountAddress,
        callData,
    };
    const userOp = await prefillUserOpWithDefaults(partialUserOp);

    // new ticket for paymaster
    const ticket = {
        sender: accountAddress,
        target: payOutAddress,
        maxCost: ethers.parseEther("100"),
        deadline: Math.floor(Date.now() / 1000) + 3600, // now + 1h
        chainId,
    };
    userOp.paymasterAndData = await buildPaymasterAndData(paymaster, pmAuthority, ticket);

    // sign full user operation with the owner of the account (FirstAccount)
    userOp.signature = await signUserOp(entryPoint, userOp, owner);
    
    // balance before
    const beforeUser = await payoutEth.getBalance(userAddress);
    const beforePM   = await entryPoint.getDeposit(paymasterAddress);

    // expect(beforeUser).to.eq(0n, "user account should start with 0 ETH");

    // execute the user operation with paymaster -> gasless for user
    // normally done by the off-chain bundler
    await entryPoint.handleOps([userOp], await beneficiary.getAddress());

    // balance after
    const afterUser = await payoutEth.getBalance(userAddress);
    const afterPM   = await entryPoint.getDeposit(paymasterAddress);

    // assert
    expect(afterUser - beforeUser).to.eq(payOutAmount, "user should receive withdrawn ETH");
    expect(afterPM).to.be.lt(beforePM, "paymaster deposit should decrease");
  });
});
