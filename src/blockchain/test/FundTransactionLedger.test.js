const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("FundTransactionLedger", function () {
  let FundTransactionLedger;
  let ledger;
  let owner;
  let addr1;
  let addr2;
  let addrs;

  beforeEach(async function () {
    // Get the ContractFactory and Signers here.
    FundTransactionLedger = await ethers.getContractFactory("FundTransactionLedger");
    [owner, addr1, addr2, ...addrs] = await ethers.getSigners();

    // Deploy a new FundTransactionLedger contract before each test
    ledger = await FundTransactionLedger.deploy();
    await ledger.deployed();

    // Grant API role to addr1
    const API_ROLE = ethers.utils.keccak256(ethers.utils.toUtf8Bytes("API_ROLE"));
    await ledger.grantRole(API_ROLE, addr1.address);
  });

  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      expect(await ledger.hasRole(await ledger.DEFAULT_ADMIN_ROLE(), owner.address)).to.equal(true);
    });

    it("Should grant API role to addr1", async function () {
      const API_ROLE = ethers.utils.keccak256(ethers.utils.toUtf8Bytes("API_ROLE"));
      expect(await ledger.hasRole(API_ROLE, addr1.address)).to.equal(true);
    });
  });

  describe("Transactions", function () {
    const testId = ethers.utils.formatBytes32String("test-id");
    const testAmount = ethers.utils.parseEther("1.0");
    const testCurrency = "KES";
    const testType = 0; // ALLOCATION
    const testDescription = "Test transaction";
    const testSourceId = ethers.utils.formatBytes32String("source-id");
    const testDestId = ethers.utils.formatBytes32String("dest-id");
    const testFundId = ethers.utils.formatBytes32String("fund-id");
    const testBudgetId = ethers.utils.formatBytes32String("budget-id");
    const testDocRef = "doc-ref";
    const testCreatedById = ethers.utils.formatBytes32String("creator-id");

    it("Should create a new transaction", async function () {
      await expect(ledger.connect(addr1).createTransaction(
        testId,
        testAmount,
        testCurrency,
        testType,
        testDescription,
        testSourceId,
        testDestId,
        testFundId,
        testBudgetId,
        testDocRef,
        testCreatedById
      )).to.emit(ledger, "TransactionCreated")
        .withArgs(testId, testType, testAmount);

      const tx = await ledger.getTransaction(testId);
      expect(tx.amount).to.equal(testAmount);
      expect(tx.currency).to.equal(testCurrency);
      expect(tx.txType).to.equal(testType);
      expect(tx.status).to.equal(0); // PENDING
    });

    it("Should approve a transaction", async function () {
      await ledger.connect(addr1).createTransaction(
        testId,
        testAmount,
        testCurrency,
        testType,
        testDescription,
        testSourceId,
        testDestId,
        testFundId,
        testBudgetId,
        testDocRef,
        testCreatedById
      );

      const approverId = ethers.utils.formatBytes32String("approver-id");
      await expect(ledger.connect(addr1).approveTransaction(testId, approverId))
        .to.emit(ledger, "TransactionApproved")
        .withArgs(testId, approverId);

      const tx = await ledger.getTransaction(testId);
      expect(tx.status).to.equal(1); // APPROVED
    });

    it("Should reject a transaction", async function () {
      await ledger.connect(addr1).createTransaction(
        testId,
        testAmount,
        testCurrency,
        testType,
        testDescription,
        testSourceId,
        testDestId,
        testFundId,
        testBudgetId,
        testDocRef,
        testCreatedById
      );

      const rejecterId = ethers.utils.formatBytes32String("rejecter-id");
      const reason = "Invalid transaction";
      await expect(ledger.connect(addr1).rejectTransaction(testId, rejecterId, reason))
        .to.emit(ledger, "TransactionRejected")
        .withArgs(testId, rejecterId, reason);

      const tx = await ledger.getTransaction(testId);
      expect(tx.status).to.equal(2); // REJECTED
    });

    it("Should complete an approved transaction", async function () {
      await ledger.connect(addr1).createTransaction(
        testId,
        testAmount,
        testCurrency,
        testType,
        testDescription,
        testSourceId,
        testDestId,
        testFundId,
        testBudgetId,
        testDocRef,
        testCreatedById
      );

      const approverId = ethers.utils.formatBytes32String("approver-id");
      await ledger.connect(addr1).approveTransaction(testId, approverId);

      await expect(ledger.connect(addr1).completeTransaction(testId))
        .to.emit(ledger, "TransactionCompleted")
        .withArgs(testId);

      const tx = await ledger.getTransaction(testId);
      expect(tx.status).to.equal(3); // COMPLETED
    });

    it("Should flag a transaction", async function () {
      await ledger.connect(addr1).createTransaction(
        testId,
        testAmount,
        testCurrency,
        testType,
        testDescription,
        testSourceId,
        testDestId,
        testFundId,
        testBudgetId,
        testDocRef,
        testCreatedById
      );

      const reason = "Suspicious activity detected";
      await expect(ledger.connect(addr1).flagTransaction(testId, reason))
        .to.emit(ledger, "TransactionFlagged")
        .withArgs(testId, reason);

      const tx = await ledger.getTransaction(testId);
      expect(tx.status).to.equal(4); // FLAGGED
      expect(tx.aiFlagged).to.equal(true);
    });
  });

  describe("Access Control", function () {
    it("Should not allow non-API role to create transaction", async function () {
      const testId = ethers.utils.formatBytes32String("test-id");
      await expect(ledger.connect(addr2).createTransaction(
        testId,
        ethers.utils.parseEther("1.0"),
        "KES",
        0,
        "Test",
        ethers.utils.formatBytes32String("source"),
        ethers.utils.formatBytes32String("dest"),
        ethers.utils.formatBytes32String("fund"),
        ethers.utils.formatBytes32String("budget"),
        "doc",
        ethers.utils.formatBytes32String("creator")
      )).to.be.revertedWith("AccessControl");
    });
  });
}); 