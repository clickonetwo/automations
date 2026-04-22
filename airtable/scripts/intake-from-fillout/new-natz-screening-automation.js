/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

// Global definitions
import { base, input } from "../library/airtable.internal";

// noinspection SpellCheckingInspection
const formNamedFieldMap = {
    masterLink: "fldZImdW90sWSc41X", // Potential Client
};

// noinspection SpellCheckingInspection
const attachmentFieldMap = {
    // map from field IDs in General Inquiry Form table to All Contacts table
    fldBnoXmLGWCTwQq2: "fldtWTfL7aLfTjMLd", // Permanent Resident Card -> Documents Provided
};

// Script invocation (as automation)
// noinspection SpellCheckingInspection
const formTable = base.getTable("tblDaONedyuHpGEQG"); // Natz Screening Form
// noinspection SpellCheckingInspection
const masterTable = base.getTable("tblsnJnJ4ubpZFLwq"); // All Contacts Master Table
const { formRecordId } = input.config();
const formRecord = await formTable.selectRecordAsync(formRecordId);
if (!formRecord) {
    throw new Error(`No screening form record exists for ID ${formRecordId}`);
}
const { masterRecordId } = await extractMatchData();
await newScreeningRecordAction();

async function extractMatchData() {
    const masterLinks = formRecord.getCellValue(formNamedFieldMap.masterLink) || [];
    if (masterLinks.length > 1) {
        throw new Error(
            "Form refers to multiple master records: " + masterLinks.map((l) => l.id).join(", ")
        );
    } else if (masterLinks.length === 0) {
        throw new Error(`No Potential Client found in form record ${formRecordId}`);
    }
    const masterRecordId = masterLinks[0].id;
    return { masterRecordId };
}

async function newScreeningRecordAction() {
    let masterRecord = await masterTable.selectRecordAsync(masterRecordId);
    if (!masterRecord) {
        throw new Error(`Form refers to master record ${masterRecordId}, but it doesn't exist`);
    }
    let count = 0;
    let masterFields = {};
    for (let [srcKey, targetKey] of Object.entries(attachmentFieldMap)) {
        const master = masterRecord.getCellValue(targetKey) || [];
        const form = formRecord.getCellValue(srcKey) || [];
        if (form.length) {
            count += 1;
            masterFields[targetKey] = [...master, ...form];
        }
    }
    if (count) {
        await masterTable.updateRecordAsync(masterRecord.id, masterFields);
    } else {
        console.warn(`Form ${formRecordId} has no attachments`);
    }
}
