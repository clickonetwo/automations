/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

const inputs = input.config()

const stringFieldMap = {  // map from field IDs in Jotform table to All Contacts table
    "fldds7G7kLLDhDpOU": "fldGF8G0cEoxqKgrd",   // Name
    "fldR38yGi44MzEzCd": "fldm2CYBimrY5o54y",   // Jotform Language Filled Out
    "fldb1XkCQuY0DRPt8": "fldli2SXrunrRmRap",   // Email
    "fldmBN0iKS0y25ARO": "fldpgAXOS3LnRVb4B",   // City
    "fld3MaagynWcu9SwU": "fldp6DqsrrskD9Kwc",   // Preferred Language (Other)
    "fldiEt86mW7RYi80I": "fldXCH3uVyuqLEzJY",   // Service Request Information
}

const choiceFieldMap = {  // map from field IDs in Jotform table to All Contacts table
    "fldEkoLa07q0uTcja": "fld0q3mP5U6GRA1uT",   // State
    "fldMnQhmMHYE8J0bd": "fld3SiQOVVKjHrw9n",   // LGBTQ+?
    "fldvD9qaA2cIIcuBI": "fldfeDUc7AwkLAcxn",   // Preferred Language
    "fldjlWXohDio3nRhD": "fldwmDDotAciBF8xK",   // Requested Legal Assistance
    "fld9TGWodwEhWPyKF": "fld5eDcXPls53cTjb",   // In Removal Proceedings
}

const fieldMap = {
    ...stringFieldMap,
    ...choiceFieldMap,
}

await newJotformRecordAction(base, inputs.recordId, inputs.usPhone, inputs.intlPhone)

async function newJotformRecordAction(base, recordId, usPhone, intlPhone) {
    let canonicalPhone
    if (usPhone) {
        canonicalPhone = usPhoneIntoE164(usPhone)
    } else {
        canonicalPhone = intlPhoneIntoE164(intlPhone)
    }
    const thisTable = base.getTable("tbldpkhtbhPAJlLd5")    // Jotform Contact Log Master Table
    const masterTable = base.getTable("tblsnJnJ4ubpZFLwq")  // All Contacts Master Table
    let masterDataFields = Object.entries(fieldMap).map(pair => pair[1])
    let result = await masterTable.selectRecordsAsync({
        fields: [
            "fld4lEBvUftT8MoGs",    // E.164 Number
            "fld4GUTSNxidFqYJf",    // Jotform Contacts from Person
            ...masterDataFields,
        ],
        sorts: [{field: "fldA1vt9a0BtASfib", direction: 'asc'}],    // Created date
    })
    let matching = result.records.filter(r => (r.getCellValue("fld4lEBvUftT8MoGs") === canonicalPhone))
    if (matching.length === 0) {
        console.log(`No master record found; creating one`);
        await makeNewJotformMasterRecord(thisTable, masterTable, recordId, canonicalPhone)
    } else if (matching.length === 1) {
        console.log(`One matching master record found; updating it`)
        await updateExistingJotformMasterRecords(thisTable, masterTable, matching, recordId)
    } else {
        console.log(`Multiple matching master records found, updating all and marking as duplicates`)
        await updateExistingJotformMasterRecords(thisTable, masterTable, matching, recordId)
        await markMasterRecordsAsDuplicates(masterTable, matching)
    }
}

async function makeNewJotformMasterRecord(thisTable, masterTable, recordId, phoneNumber) {
    const thisRecord = (await thisTable.selectRecordsAsync({
        recordIds: [recordId],
        fields: Object.entries(fieldMap).map(pair => pair[0])
    })).records[0]
    const targetRecordFields = {}
    for (let [srcKey, targetKey] of Object.entries(stringFieldMap)) {
        targetRecordFields[targetKey] = thisRecord.getCellValueAsString(srcKey)
    }
    for (let [srcKey, targetKey] of Object.entries(choiceFieldMap)) {
        const val = thisRecord.getCellValueAsString(srcKey)
        if (val) {
            targetRecordFields[targetKey] = {name: val}
        }
    }
    targetRecordFields["fld4lEBvUftT8MoGs"] = phoneNumber,               // E.164 Number
    targetRecordFields["fld4GUTSNxidFqYJf"] = [{id: recordId}],          // Jotform Contacts from Person
    targetRecordFields["fld1CNjHs3PRuqCok"] = formatPhone(phoneNumber),  // Phone
    await masterTable.createRecordAsync(targetRecordFields)
}

async function updateExistingJotformMasterRecords(thisTable, masterTable, masterRecords, newRecordId) {
    const thisRecord = (await thisTable.selectRecordsAsync({
        recordIds: [newRecordId],
        fields: Object.entries(fieldMap).map(pair => pair[0])
    })).records[0]
    const updates = []
    for (const masterRecord of masterRecords) {
        let existingLinks = masterRecord.getCellValue("fld4GUTSNxidFqYJf")  // Jotform Contacts from Person
        if (existingLinks) {
            if (existingLinks && existingLinks.map(v => v.id).includes(newRecordId)) {
                console.log(`Master record ${masterRecord.id} is already linked to this form; skipping update`)
                continue
            }
            existingLinks.push({id: newRecordId})
        } else {
            existingLinks = [{id: newRecordId}]
        }
        const targetRecordFields = {"fld4GUTSNxidFqYJf": existingLinks}
        for (let [srcKey, targetKey] of Object.entries(stringFieldMap)) {
            const val = masterRecord.getCellValue(targetKey)
            if (!val) {
                targetRecordFields[targetKey] = thisRecord.getCellValueAsString(srcKey)
            }
        }
        for (let [srcKey, targetKey] of Object.entries(choiceFieldMap)) {
            const mval = masterRecord.getCellValue(targetKey)
            if (!mval) {
                const val = thisRecord.getCellValueAsString(srcKey)
                if (val) {
                    targetRecordFields[targetKey] = {name: val}
                }
            }
        }
        updates.push({id: masterRecord.id, fields: targetRecordFields})
    }
    for (let i = 0; i < updates.length; i += 50) {
        const end = Math.min(updates.length, i + 50)
        await masterTable.updateRecordsAsync(updates.slice(i, end))
    }
}

async function markMasterRecordsAsDuplicates(masterTable, masterRecords) {
    const updates = masterRecords.map((r) => ({
        id: r.id, fields: {
            "fldEVYjKOxyLSYJZF": true   // Has Duplicates?
        }
    }))
    for (let i = 0; i < updates.length; i += 50) {
        const end = Math.min(updates.length, i + 50)
        await masterTable.updateRecordsAsync(updates.slice(i, end))
    }
}

// takes a phone in E.164 format and formats it for display
function formatPhone(phone) {
    const twoDigitCountryCodes = [
        "20", "27", "30", "31", "32", "33", "34", "36", "39",
        "40", "41", "43", "44", "45", "46", "47", "48", "49",
        "51", "52", "53", "54", "55", "56", "57", "58",
        "60", "61", "62", "63", "64", "65", "66",
        "70", "71", "72", "73", "74", "75", "76", "77", "78", "79",
        "81", "82", "84", "86", "87", "88",
        "90", "91", "92", "93", "94", "95", "98",
    ]
    if (phone.startsWith("+1")) {
        // Zone 1
        return `(${phone.substring(2,5)}) ${phone.substring(5,8)}-${phone.substring(8,12)}`;
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
        // put last four numbers together
        suffixSuffix = "-" + suffix.substring(suffix.length - 4);
        suffix = suffix.substring(0, suffix.length - 4);
    }
    for (let i = suffix.length; i > 3; i = i - 3) {
        suffix = suffix.substring(0, i-3) + "-" + suffix.substring(i-3)
    }
    return prefix + "-" + suffix + suffixSuffix;
}

// canonicalize US phone number into E.164 format
function usPhoneIntoE164(phone) {
    let digits = phone.replace(/\D/g,'')
    if (digits.length === 10) {
        return "+1" + digits
    }
    if (digits.length === 11 && digits.charAt(0) === "1") {
        return "+" + digits
    }
    // yuk - this is a pretty strange phone number
    // just return a place-holder so we group all
    // the duplicates together
    return "+01112223333"
}

// canonicalize international phone number into E.164 format
function intlPhoneIntoE164(phone) {
    let digits = phone.replace(/\D/g,'')
    if (digits.startsWith("001")) {
        // strip the international dialing prefix
        digits = digits.substring(3)
    }
    while (digits.charAt(0) === "0") {
        digits = digits.substring(1)
    }
    if (digits.length < 8) {
        // not a valid number, return place holder
        return "+009998887777"
    }
    if (digits.charAt(0) === "1") {
        // this is a Zone 1 number - perhaps they are in the Caribbean?
        if (digits.length === 11) {
            return "+" + digits
        }
        // not a valid international number
        return "+009998887777"
    }
    return "+" + digits
}
