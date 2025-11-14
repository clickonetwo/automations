/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

await updateInternationalPhoneExtension()

async function updateInternationalPhoneExtension() {
    const masterTable = base.getTable("tblsnJnJ4ubpZFLwq")
    const found = await findRecordsWithInternationalPhones(masterTable)
    await updateInternationalPhones(masterTable, found)
}

async function findRecordsWithInternationalPhones(masterTable) {
    output.text(`Finding records with international phones...`)
    let result = await masterTable.selectRecordsAsync({
        fields: ["fld4lEBvUftT8MoGs"],    // E.164 number
    })
    let records = result.records.map(r => {
        const e164 = r.getCellValueAsString("fld4lEBvUftT8MoGs")
        return {id: r.id, phone: e164}
    })
    records = records.filter(r => r.phone && !r.phone.startsWith("+1"))
    output.text(`Found ${records.length} records to update.`)
    return records
}

async function updateInternationalPhones(masterTable, records) {
    const updates = []
    for (const record of records) {
        const phone = formatPhone(record.phone)
        const update = {
            id: record.id,
            fields: {"fld1CNjHs3PRuqCok": phone},   // Phone
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
