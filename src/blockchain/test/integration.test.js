const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("Digital Watchdog Integration Tests", function () {
  let transactionLogger;
  let fundManager;
  let owner;
  let apiService;
  let user1;
  let user2;

  beforeEach(async function () {
    // Get signers
    [owner, apiService, user1, user2] = await ethers.getSigners();

    // Deploy TransactionEventLogger
    const TransactionEventLogger = await ethers.getContractFactory("TransactionEventLogger");
    transactionLogger = await TransactionEventLogger.deploy();
    await transactionLogger.deployed();

    // Deploy FundManager
    const FundManager = await ethers.getContractFactory("FundManager");
    fundManager = await FundManager.deploy();
    await fundManager.deployed();

    // Grant API_ROLE to apiService
    const API_ROLE = ethers.utils.keccak256(ethers.utils.toUtf8Bytes("API_ROLE"));
    await transactionLogger.grantRole(API_ROLE, apiService.address);

    // Add apiService as authorized updater in FundManager
    await fundManager.addAuthorizedUpdater(apiService.address);
  });

  describe("Fund Management Flow", function () {
    it("should create a fund and manage transactions", async function () {
      // Create a fund
      const fundId = ethers.utils.formatBytes32String("FUND001");
      const fundName = "Education Fund 2024";
      const fundCode = "EDU24";
      const fundCategory = "Education";
      const fiscalYear = "2024";
      const totalAmount = ethers.utils.parseEther("1000"); // 1000 ETH equivalent

      await fundManager.connect(apiService).createFund(
        fundId,
        fundName,
        fundCode,
        fundCategory,
        fiscalYear,
        totalAmount
      );

      // Verify fund creation
      const fund = await fundManager.funds(fundId);
      expect(fund.name).to.equal(fundName);
      expect(fund.totalAmount).to.equal(totalAmount);

      // Create a transaction
      const txId = ethers.utils.formatBytes32String("TX001");
      const amount = ethers.utils.parseEther("100");
      const currency = "KES";
      const txType = 0; // ALLOCATION
      const description = "Initial allocation";
      const sourceId = ethers.utils.formatBytes32String("SRC001");
      const destinationId = ethers.utils.formatBytes32String("DST001");
      const budgetLineItemId = ethers.utils.formatBytes32String("BUD001");
      const documentRef = "ipfs://QmDocument001";
      const createdById = ethers.utils.formatBytes32String("USER001");

      await fundManager.connect(apiService).createTransaction(
        txId,
        amount,
        currency,
        txType,
        description,
        sourceId,
        destinationId,
        fundId,
        budgetLineItemId,
        documentRef,
        createdById
      );

      // Log transaction creation event
      const creatorHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(createdById));
      const detailsHash = ethers.utils.keccak256(
        ethers.utils.defaultAbiCoder.encode(
          ["uint256", "string", "uint8", "bytes32", "bytes32", "bytes32", "bytes32"],
          [amount, currency, txType, sourceId, destinationId, fundId, budgetLineItemId]
        )
      );

      await transactionLogger.connect(apiService).recordTransactionCreation(
        txId,
        creatorHash,
        detailsHash
      );

      // Verify transaction creation
      const tx = await fundManager.transactions(txId);
      expect(tx.amount).to.equal(amount);
      expect(tx.status).to.equal(0); // PENDING

      // Get transaction events
      const events = await transactionLogger.getTransactionEvents(txId);
      expect(events.eventTypes[0]).to.equal(0); // CREATED
      expect(events.actorHashes[0]).to.equal(creatorHash);
      expect(events.detailsHashes[0]).to.equal(detailsHash);

      // Approve transaction
      const approverId = ethers.utils.formatBytes32String("APPROVER001");
      const approverHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(approverId));

      await fundManager.connect(apiService).approveTransaction(txId, approverId);
      await transactionLogger.connect(apiService).recordApproval(txId, approverHash);

      // Verify approval
      const updatedTx = await fundManager.transactions(txId);
      expect(updatedTx.status).to.equal(1); // APPROVED

      // Complete transaction
      const completerId = ethers.utils.formatBytes32String("COMPLETER001");
      const completerHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(completerId));

      await fundManager.connect(apiService).completeTransaction(txId, completerId);
      await transactionLogger.connect(apiService).recordCompletion(txId, completerHash);

      // Verify completion and fund balance update
      const finalTx = await fundManager.transactions(txId);
      expect(finalTx.status).to.equal(3); // COMPLETED

      const updatedFund = await fundManager.funds(fundId);
      expect(updatedFund.allocated).to.equal(amount);
    });

    it("should handle AI flagging of suspicious transactions", async function () {
      // Create a fund
      const fundId = ethers.utils.formatBytes32String("FUND002");
      await fundManager.connect(apiService).createFund(
        fundId,
        "Emergency Fund",
        "EMG24",
        "Emergency",
        "2024",
        ethers.utils.parseEther("1000")
      );

      // Create a suspicious transaction
      const txId = ethers.utils.formatBytes32String("TX002");
      await fundManager.connect(apiService).createTransaction(
        txId,
        ethers.utils.parseEther("500"),
        "KES",
        1, // DISBURSEMENT
        "Emergency supplies",
        ethers.utils.formatBytes32String("SRC002"),
        ethers.utils.formatBytes32String("DST002"),
        fundId,
        ethers.utils.formatBytes32String("BUD002"),
        "ipfs://QmDocument002",
        ethers.utils.formatBytes32String("USER002")
      );

      // Flag transaction
      const flagReason = "Unusually large disbursement amount";
      await fundManager.connect(apiService).flagTransaction(txId, flagReason);
      await transactionLogger.connect(apiService).recordFlagging(txId, flagReason);

      // Verify flagging
      const tx = await fundManager.transactions(txId);
      expect(tx.status).to.equal(4); // FLAGGED
      expect(tx.aiFlagged).to.be.true;
      expect(tx.aiFlagReason).to.equal(flagReason);

      // Verify event log
      const events = await transactionLogger.getTransactionEvents(txId);
      expect(events.eventTypes).to.include(4); // FLAGGED
      expect(events.metadatas[events.eventTypes.length - 1]).to.equal(flagReason);
    });
  });
}); 