// noinspection SpellCheckingInspection

/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

import { base, input } from "../library/airtable.internal";
import { GbTransactionPayload } from "./payloads";

const gbTransactionsTable = base.getTable("tblAjtXdKANH06DIn"); // GB Transactions
const gbPlansTable = base.getTable("tblqgeY56ThyqHjIM"); // GB Plans
const contactsTable = base.getTable("tblrsTMxY82X7DyZG"); // Contacts
const donationsTable = base.getTable("tblUoq0LQPx0WEGHf"); // Donations
const campaignsTable = base.getTable("tbltHW0T3EbgPX3Ff"); // Campaigns

const gbTransactionsPayloadFieldId = "fldxlf48Q1rsQ41XM"; // Payload
const gbTransactionsIdFieldId = "fldMw3mEFuRy5dAG4"; // Transaction ID

const gbPlansGbIdFieldId = "fld345XhcHaQV5Oaj"; // GiveButter ID
const gbPlansStatusFieldId = "fld70U96Fh0erAd27";
const gbPlansFrequencyFieldId = "fldPF06oq9X6h1MMn";
const gbPlansStartedFieldId = "fldxDW7rQGrV9iQGd";
const gbPlansEndedFieldId = "fldZk7rhKQxmBIxin";
const gbPlansDonationsFieldId = "fldXV9rVvyhMjXK4M"; // Donations

const donationsGbIdFieldId = "fldHvNcrKLuNvszQV"; // GiveButter ID
const donationsTypeFieldId = "fldUYFpeaQKOvtKcn"; // Donation Type (`Donation`)
const donationsAmountFieldId = "fldGJVoeZk9alsayT"; // Donation Amount
const donationsCampaignsFieldId = "fldhYecexSdiuFN8O"; // Campaigns
const donationsDateFieldId = "fld6ILU3ZDzbAHjRb"; // Paid Donation Date
const donationsStatusFieldId = "fldvq258bMN1EnZrS"; // Donation Status (`Paid`)
const donationsSourceFieldId = "fldMlp1KAU2O5aDzX"; // Donation Source (`GiveButter - EFT`)
const donationsTransactionFieldId = "fldNjF3cO8eeJHven"; // GiveButter Transaction
const donationsRecurringFieldId = "fldkHqyj8d8iJzmwd"; // Recurring?

const contactsGbIdFieldId = "fldqqPq1b9b7VOVey"; // GiveButter ID
const contactsFirstNameFieldId = "fldlQ6XJ3xd1tQU1x"; // First Name
const contactsLastNameFieldId = "fld7BcJ3eYGn0xCWO"; // Last Name
const contactsEmail1FieldId = "fldjGQciq7ccIRfsm"; // Email1
const contactsEmail2FieldId = "fldGHRReWTPk49t1Y"; // Email2
const contactsAddressFieldId = "fldgcYg5x3zQOUggS"; // Street Address
const contactsCityFieldId = "fldi2SAXzvfOOQy4W"; // City
const contactsStateFieldId = "fldD0In6imjjhtiGL"; // State
const contactsZipFieldId = "fld97L7HdcYpviw5P"; // Zip
const contactsCountryFieldId = "fldtoXiM2ehKs68PS"; // Country
const contactsPhoneFieldId = "fldHbJGJBn5hVpzeU"; // Phone # (formatted)
const contactsDonationsFieldId = "fldZfXELOVTAGf5Mc"; // Donations Made

const campaignsNameFieldId = "fld6wji2XyoNolXUt"; // Campaign Name
const campaignsGbIdFieldId = "fldcP7RwCM6Ht9K2S"; // GiveButter ID
const campaignsGbCodeFieldId = "fldnVTTDXBJKoiyI5"; // GiveButter Code
const campaignsStartDateFieldId = "fldFjj6HzcfXekJx0"; // Start Date

const { transactionRecordId } = input.config();
await processNewTransaction(transactionRecordId);

async function processNewTransaction(recordId) {
    let data = await validateAndLabelTransaction(recordId);
    let donationRecordId = await createOrUpdateDonation(data, recordId);
    if (data.plan_id) {
        await createOrUpdatePlan(data, donationRecordId);
    }
    await createOrUpdateDonor(data, donationRecordId);
    await maybeSubscribeDonorToNewsletter(data);
}

/**
 * Validates the transaction record and labels it (with the transaction ID) as processed.
 * @param {string} recordId
 * @returns {Promise<GbTransactionData>}
 */
async function validateAndLabelTransaction(recordId) {
    const transactionRecord = await gbTransactionsTable.selectRecordAsync(recordId);
    if (!transactionRecord) {
        throw new Error(`No transaction record exists for record ID ${transactionRecordId}`);
    }
    /** @type {GbTransactionPayload} */
    const payload = JSON.parse(
        transactionRecord.getCellValueAsString(gbTransactionsPayloadFieldId)
    );
    if (!payload) {
        throw new Error(`No payload data for transaction record ${transactionRecordId}`);
    }
    // see if we have already processed this transaction event
    const data = payload.data;
    const id = data.id;
    const existing = await gbTransactionsTable.selectRecordsAsync({
        fields: [gbTransactionsIdFieldId],
    });
    const matchingRecords = existing.records.filter(
        (r) => r.getCellValue(gbTransactionsIdFieldId) === id
    );
    if (matchingRecords.length) {
        if (matchingRecords.length > 1) {
            throw new Error(`There are multiple existing transaction records with id ${id}`);
        } else if (matchingRecords[0].id !== transactionRecordId) {
            throw new Error(
                `This transaction was already processed in record ${matchingRecords[0].id}`
            );
        } else {
            console.warn(`Transaction record ${transactionRecordId} is being reprocessed`);
        }
    } else {
        console.log(`Transaction record ${transactionRecordId} is valid and being processed`);
        await gbTransactionsTable.updateRecordAsync(transactionRecordId, {
            [gbTransactionsIdFieldId]: id,
        });
    }
    return data;
}

/**
 * Creates (or updates) the donation record for a transaction and returns its record ID.
 * @param {GbTransactionData} data
 * @param {string} transactionRecordId
 * @returns {Promise<string>}
 */
async function createOrUpdateDonation(data, transactionRecordId) {
    let recordId = "";
    const existing = await donationsTable.selectRecordsAsync({
        fields: [donationsGbIdFieldId],
    });
    const matchingRecords = existing.records.filter(
        (r) => r.getCellValue(donationsGbIdFieldId) === data.id
    );
    if (matchingRecords.length) {
        if (matchingRecords.length > 1) {
            throw new Error(`There is more than one donation record with ID ${data.id}`);
        }
        console.log(`Reprocessing existing donation record ${matchingRecords[0].id}`);
        recordId = matchingRecords[0].id;
    } else {
        console.log(`Creating new donation record for transaction ${data.id}`);
        recordId = await donationsTable.createRecordAsync({
            [donationsGbIdFieldId]: data.id,
            [donationsTransactionFieldId]: [{ id: transactionRecordId }],
            [donationsTypeFieldId]: { name: "Donation" },
            [donationsStatusFieldId]: { name: "Paid" },
            [donationsSourceFieldId]: { name: "GiveButter - EFT" },
        });
    }
    console.log(`Updating donation record ${recordId} with transaction data`);
    const campaignId = data.campaign_id.toString();
    const fieldUpdates = {
        [donationsAmountFieldId]: data.amount,
        [donationsCampaignsFieldId]: [{ id: await createOrUpdateCampaign(campaignId) }],
        [donationsDateFieldId]: data.created_at,
        [donationsRecurringFieldId]: data.plan_id !== null,
    };
    await donationsTable.updateRecordAsync(recordId, fieldUpdates);
    return recordId;
}

/**
 * Creates (or updates) the campaign record for a transaction and returns its record ID.
 * @param {string} campaignId
 * @returns {Promise<string>}
 */
async function createOrUpdateCampaign(campaignId) {
    const existing = await campaignsTable.selectRecordsAsync({
        fields: [campaignsGbIdFieldId],
    });
    const matchingRecords = existing.records.filter(
        (r) => r.getCellValue(campaignsGbIdFieldId) === campaignId
    );
    if (matchingRecords.length) {
        if (matchingRecords.length > 1) {
            throw new Error(`There is more than one campaign record with ID ${campaignId}`);
        }
        console.log(`Found existing campaign record ${matchingRecords[0].id}`);
        return matchingRecords[0].id;
    }
    console.log(`Creating new campaign record for campaign ${campaignId}`);
    let data = await fetchCampaign(campaignId);
    let id = await campaignsTable.createRecordAsync({
        [campaignsGbIdFieldId]: campaignId,
        [campaignsGbCodeFieldId]: data.code ?? "",
        [campaignsNameFieldId]: data.title,
        [campaignsStartDateFieldId]: data.created_at.slice(0, 10),
    });
    await notifyNew(
        "campaign",
        ["leanne.brotsky@oasislegalservices.org", "daniel.brotsky@oasislegalservices.org"],
        [
            `Campaign title: ${data.title}`,
            `Campaign ID: ${campaignId}`,
            `Campaign code: ${data.code ?? "(none)"}`,
        ]
    );
    return id;
}

/**
 * Creates (or updates) the donor record for a transaction and returns its record ID.
 * @param {GbTransactionData} data
 * @param {string} donationRecordId
 * @returns {Promise<string>}
 */
async function createOrUpdateDonor(data, donationRecordId) {
    const donorId = data.contact_id.toString();
    const existing = await contactsTable.selectRecordsAsync({
        fields: [
            contactsGbIdFieldId,
            contactsEmail1FieldId,
            contactsEmail2FieldId,
            contactsDonationsFieldId,
        ],
    });
    // check for matching contacts by GiveButter ID, email1, or email2
    let matchingRecords = existing.records.filter(
        (r) => r.getCellValue(contactsGbIdFieldId) === donorId
    );
    if (!matchingRecords.length) {
        matchingRecords = existing.records.filter(
            (r) => r.getCellValue(contactsEmail1FieldId) === data.email
        );
    }
    if (!matchingRecords.length) {
        matchingRecords = existing.records.filter(
            (r) => r.getCellValue(contactsEmail2FieldId) === data.email
        );
    }
    // if there is a matching contact, make sure this donation is linked
    if (matchingRecords.length) {
        if (matchingRecords.length > 1) {
            throw new Error(`There is more than one donor with ID ${donorId}`);
        }
        const donorRecord = matchingRecords[0];
        const existing = donorRecord.getCellValue(contactsDonationsFieldId);
        if (existing && !existing.map((v) => v.id).includes(donationRecordId)) {
            console.log(`This donation link is being added to existing contact ID ${donorId}`);
            const fieldUpdates = {
                [contactsDonationsFieldId]: [...existing, { id: donationRecordId }],
            };
            // also make sure that the GB ID and emails for this contact are up to date
            const gbId = donorRecord.getCellValue(contactsGbIdFieldId);
            const email1 = donorRecord.getCellValue(contactsEmail1FieldId);
            const email2 = donorRecord.getCellValue(contactsEmail2FieldId);
            if (!gbId) {
                fieldUpdates[contactsGbIdFieldId] = donorId;
            }
            if (!email1) {
                fieldUpdates[contactsEmail1FieldId] = data.email;
            } else if (email1 !== data.email && !email2) {
                fieldUpdates[contactsEmail2FieldId] = data.email;
            }
            await contactsTable.updateRecordAsync(donorRecord.id, fieldUpdates);
        } else {
            console.log(`This donation is already linked to existing contact ID ${donorId}`);
        }
        return donorRecord.id;
    }
    // if there is no matching contact, create one and link this donation
    console.log(`Contact ${donorId} is being added and linked to this donation`);
    const newFields = {
        [contactsGbIdFieldId]: donorId,
        [contactsFirstNameFieldId]: data.first_name,
        [contactsLastNameFieldId]: data.last_name,
        [contactsEmail1FieldId]: data.email,
        [contactsAddressFieldId]:
            data.address.address_1 + (data.address.address_2 ? "\n" + data.address.address_2 : ""),
        [contactsCityFieldId]: data.address.city,
        [contactsStateFieldId]: data.address.state,
        [contactsZipFieldId]: data.address.zipcode,
        [contactsCountryFieldId]: data.address.country,
        [contactsPhoneFieldId]: data.phone ? formatPhone(data.phone) : "",
        [contactsDonationsFieldId]: [{ id: donationRecordId }],
    };
    let id = await contactsTable.createRecordAsync(newFields);
    await notifyNew(
        "contact",
        ["leanne.brotsky@oasislegalservices.org", "daniel.brotsky@oasislegalservices.org"],
        [
            `Contact Name: ${data.first_name} ${data.last_name}`,
            `Contact Email: ${data.email}`,
            `Contact ID: ${donorId}`,
        ]
    );
    return id;
}

/**
 * Creates (or updates) the plan record for a transaction and returns its record ID.
 * @param {GbTransactionData} data
 * @param {string} donationRecordId
 * @returns {Promise<string>} The record ID of the created or updated plan
 */
async function createOrUpdatePlan(data, donationRecordId) {
    const planId = data.plan_id.toString();
    const existing = await gbPlansTable.selectRecordsAsync({
        fields: [gbPlansGbIdFieldId, gbPlansDonationsFieldId],
    });
    const matchingRecords = existing.records.filter(
        (r) => r.getCellValue(gbPlansGbIdFieldId) === planId
    );
    // if there is a matching plan, make sure this donation is linked
    if (matchingRecords.length) {
        if (matchingRecords.length > 1) {
            throw new Error(`There is more than one plan with ID ${planId}`);
        }
        const planRecord = matchingRecords[0];
        const existing = planRecord.getCellValue(gbPlansDonationsFieldId);
        if (existing && !existing.map((v) => v.id).includes(donationRecordId)) {
            console.log(`Donation is being added to existing plan ${planId}`);
            await gbPlansTable.updateRecordAsync(planRecord.id, {
                [gbPlansDonationsFieldId]: [...existing, { id: donationRecordId }],
            });
        } else {
            console.log(`Donation is already linked to existing plan ${planId}`);
        }
        return planRecord.id;
    }
    // otherwise fetch and create the matching plan
    console.log(`Plan ${planId} is being fetched and linked to this donation`);
    const payload = await fetchPlan(planId);
    const fieldValues = {
        [gbPlansGbIdFieldId]: planId,
        [gbPlansDonationsFieldId]: [{ id: donationRecordId }],
        [gbPlansStatusFieldId]: payload.status,
        [gbPlansFrequencyFieldId]: payload.frequency,
        [gbPlansStartedFieldId]: payload.start_at.slice(0, 10),
    };
    if (payload.canceled_at) {
        fieldValues[gbPlansEndedFieldId] = payload.canceled_at.slice(0, 10);
    }
    return await gbPlansTable.createRecordAsync(fieldValues);
}

/**
 * Sends an email to the given recipients notifying them of the new record type.
 * @param {string} recordType
 * @param {string[]} recipients
 * @param {string[]} details
 * @returns {Promise<void>}
 */
async function notifyNew(recordType, recipients, details) {
    if (!recordType || !recipients.length || !details.length) {
        throw new Error(`Invalid parameters for ${recordType} notification`);
    }
    const subject = `GiveButter donation created a new ${recordType}`;
    let body = `<p>A new ${recordType} was created while processing a GiveButter donation.</p>\n`;
    body += `<p>The details are:</p>\n`;
    body += `<ul>\n${details.map((d) => `  <li>${d}</li>`).join("\n")}</ul>\n`;
    await sendEmail(recipients, subject, body);
}

/**
 * Sends an email to the given recipients with the given subject and body.
 * @param {string[]} recipients
 * @param {string} subject
 * @param {string} body
 * @returns {Promise<void>}
 */
async function sendEmail(recipients, subject, body) {
    const apiKey = "cugXQDNfMdsfjmjAgngxumqWs";
    const endpoint = "https://hook.us1.make.com/opm6qmm3m5kgdborpy7dmva96ftne1dg";
    const response = await fetch(endpoint, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-Make-Apikey": apiKey,
        },
        body: JSON.stringify({
            to: recipients,
            subject: subject,
            body: body,
        }),
    });
    if (!response.ok) {
        throw new Error(`Failed to send email: ${response.status} ${response.statusText}`);
    }
}

/**
 * If the donor has opted in to newsletter subscriptions, subscribes them to the newsletter.
 * @param {GbTransactionData} data
 * @returns {Promise<void>}
 */
async function maybeSubscribeDonorToNewsletter(data) {
    let fields = data.custom_fields;
    for (const field of fields) {
        if (field.title.includes("subscribe to the") && field.value) {
            console.log(`Subscribing donor ${data.email} to the Oasis newsletter`);
            const apiKey = "cugXQDNfMdsfjmjAgngxumqWs";
            const endpoint = "https://hook.us1.make.com/jl1mc3yl1qck7quodpelpy8vi32uh4iz";
            const response = await fetch(endpoint, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "X-Make-Apikey": apiKey,
                },
                body: JSON.stringify({
                    email: data.email,
                    firstName: data.first_name,
                    lastName: data.last_name,
                }),
            });
            if (!response.ok) {
                throw new Error(
                    `Failed to subscribe donor ${data.email} to newsletter: ${response.status} ${response.statusText}`
                );
            }
        }
    }
}

/**
 * fetches a campaign from GiveButter's API
 * @param campaignId
 * @returns {Promise<GbCampaignData>}
 */
async function fetchCampaign(campaignId) {
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    const gbApiEndpoint = `https://api.givebutter.com/v1/campaigns/${campaignId}`;
    const response = await fetch(gbApiEndpoint, {
        method: "GET",
        headers: {
            Accept: "application/json",
            Authorization: `Bearer ${gbApiKey}`,
        },
    });
    if (!response.ok) {
        throw new Error(
            `Failed to fetch campaign ${campaignId}: ${response.status} ${response.statusText}`
        );
    }
    return await response.json();
}

/**
 * fetches a plan from GiveButter's API
 * @param planId
 * @returns {Promise<GbPlanData>}
 */
async function fetchPlan(planId) {
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    const gbApiEndpoint = `https://api.givebutter.com/v1/plans/${planId}`;
    const response = await fetch(gbApiEndpoint, {
        method: "GET",
        headers: {
            Accept: "application/json",
            Authorization: `Bearer ${gbApiKey}`,
        },
    });
    if (!response.ok) {
        throw new Error(JSON.stringify(response));
    }
    return await response.json();
}

// takes a phone in E.164 format and formats it for display
function formatPhone(phone) {
    const twoDigitCountryCodes = [
        "20",
        "27",
        "30",
        "31",
        "32",
        "33",
        "34",
        "36",
        "39",
        "40",
        "41",
        "43",
        "44",
        "45",
        "46",
        "47",
        "48",
        "49",
        "51",
        "52",
        "53",
        "54",
        "55",
        "56",
        "57",
        "58",
        "60",
        "61",
        "62",
        "63",
        "64",
        "65",
        "66",
        "70",
        "71",
        "72",
        "73",
        "74",
        "75",
        "76",
        "77",
        "78",
        "79",
        "81",
        "82",
        "84",
        "86",
        "87",
        "88",
        "90",
        "91",
        "92",
        "93",
        "94",
        "95",
        "98",
    ];
    if (phone.startsWith("+1")) {
        // Zone 1
        if (phone.length < 12) {
            return "invalid";
        }
        const part1 = `(${phone.substring(2, 5)}) ${phone.substring(5, 8)}-${phone.substring(8, 12)}`;
        if (phone.length === 12) {
            return part1;
        } else {
            return part1 + " x" + phone.substring(12);
        }
    }
    // international
    let prefix = phone.substring(0, 3);
    let suffix = phone.substring(3);
    if (!twoDigitCountryCodes.includes(phone.substring(1, 3))) {
        prefix = phone.substring(0, 4);
        suffix = phone.substring(4);
    }
    let suffixSuffix = "";
    if (suffix.length % 3 === 1) {
        // put the last four numbers together
        suffixSuffix = "-" + suffix.substring(suffix.length - 4);
        suffix = suffix.substring(0, suffix.length - 4);
    }
    for (let i = suffix.length; i > 3; i = i - 3) {
        suffix = suffix.substring(0, i - 3) + "-" + suffix.substring(i - 3);
    }
    return prefix + "-" + suffix + suffixSuffix;
}
