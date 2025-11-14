/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */


await fillE164Phones()

async function fillE164Phones() {
    const masterTable = base.getTable("tblsnJnJ4ubpZFLwq")
    const found = await findRecordsWithPhoneButNotE164Phone(masterTable)
    await updateE164Phones(masterTable, found)
}

async function findRecordsWithPhoneButNotE164Phone(masterTable) {
    output.text(`Finding records with phones but not E.164 phones...`)
    let result = await masterTable.selectRecordsAsync({
        fields: [
            "fld1CNjHs3PRuqCok",    // Phone
            "fld4lEBvUftT8MoGs",    // E.164 number
        ],
    })
    let records = result.records.map(r => {
        const phone = r.getCellValueAsString("fld1CNjHs3PRuqCok")
        const e164 = r.getCellValueAsString("fld4lEBvUftT8MoGs")
        return {id: r.id, phone, e164}
    })
    records = records.filter(r => r.phone !== "" && r.e164 === "")
    output.text(`Found ${records.length} records to update.`)
    return records
}

async function updateE164Phones(masterTable, records) {
    const updates = []
    for (const record of records) {
        const e164 = anyPhoneIntoE164(record.phone)
        if (e164.indexOf("+0") >= 0) {
            // couldn't format phone
            continue
        }
        const update = {
            id: record.id,
            fields: {"fld4lEBvUftT8MoGs": e164},   // Phone
        }
        updates.push(update)
    }
    for (let i = 0; i < updates.length; i += 50) {
        const end = Math.min(updates.length, i + 50)
        output.text(`Processing updates ${i+1} to ${end}...`)
        await masterTable.updateRecordsAsync(updates.slice(i, end))
    }
    if (updates.length > 0) {
        output.text(`Processed ${updates.length} update${updates.length === 1 ? "" : "s"}.`)
    }
}

// DON'T EDIT HERE, EDIT IN phones.js
function anyPhoneIntoE164(phone) {
    if (phone.indexOf("+") >= 0 && phone.indexOf("+1") === -1) {
        return intlPhoneIntoE164(phone)
    }
    return usPhoneIntoE164(phone)
}

// DON'T EDIT HERE, EDIT IN phones.js
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

// DON'T EDIT HERE, EDIT IN phones.js
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
        // not a valid number, return placeholder
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
