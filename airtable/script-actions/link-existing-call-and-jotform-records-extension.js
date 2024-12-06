/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

await linkExistingCallAndJotformRecords()

async function linkExistingCallAndJotformRecords() {
    const masterTable = base.getTable("tblsnJnJ4ubpZFLwq"); // All Contacts Master Table
    const map = await makePhoneToRecordMap(masterTable)
    await addCallRecordsToMasterTable(masterTable, map)
    await addFormRecordsToMasterTable(masterTable, map)
}

async function makePhoneToRecordMap(masterTable) {
    output.text(`Making map of known phone numbers to records...`)
    const map = new Map();
    let result = await masterTable.selectRecordsAsync({
        fields: [
            "fld4lEBvUftT8MoGs",    // E.164 number
            "flden5oBfu9Gniz2P",    // Dialpad Contacts from Person
            "fld4GUTSNxidFqYJf",    // Jotpad Contacts from Person
        ]
    })
    for (const record of result.records) {
        const phone = record.getCellValueAsString("fld4lEBvUftT8MoGs");
        if (phone) {
            if (map.has(phone)) {
                output.text(`Warning: E.164 number ${phone} appears on multiple records.`)
                continue
            }
            map.set(phone, record);
        }
    }
    output.text(`Found ${map.size} different phone numbers...`)
    return map;
}

async function addCallRecordsToMasterTable(masterTable, map) {
    const fromTable = base.getTable("tble5MiE5cPmRWgP7")   // Dialpad Contact Log Master Table
    const fromFieldId = "fldQEhjDhapc1tVMq"         // Phone Number
    const toFieldId = "flden5oBfu9Gniz2P"           // Dialpad Contacts from Person
    output.text(`Finding call records to be linked to the master table...`)
    const result = await fromTable.selectRecordsAsync({
        fields: [fromFieldId],
    })
    const idsAndPhones = result.records.map(r => ({id: r.id, phone: r.getCellValueAsString(fromFieldId)}))
    const records = idsAndPhones.filter(r => map.has(r.phone))
    output.text(`Found ${records.length} call records to be linked to the master table.`)
    for (let i = 0; i < records.length; i += 50) {
        output.text(`Processing updates in batch ${(i/50)+1}...`)
        const updates = []
        for (let j = i; j < records.length && j < 50; j++) {
            const record = records[j]
            const master = map.get(record.phone)
            let existingLinks = master.getCellValue(toFieldId)
            if (existingLinks) {
                if (existingLinks && existingLinks.map(v => v.id).includes(record.id)) {
                    // already linked, skip this update
                    continue
                }
                existingLinks.push({id: record.id})
            } else {
                existingLinks = [{id: record.id}]
            }
            const update = {
                id: master.id,
                fields: {[toFieldId]: existingLinks},
            }
            updates.push(update)
        }
        await masterTable.updateRecordsAsync(updates)
    }
    output.text(`Processed ${records.length} updates.`)
}

async function addFormRecordsToMasterTable(masterTable, map) {
    const fromTable = base.getTable("tbldpkhtbhPAJlLd5");  // Jotform Contact Log Master Table
    let toFieldId = "fld4GUTSNxidFqYJf"             // Jotpad Contacts from Person
    output.text(`Finding form records to be linked to the master table...`)
    const result = await fromTable.selectRecordsAsync({
        fields: [
            "fldUB3ydfv356DV2c",    // US Phone Number
            "fldvs0PPxIvoIWPzL",    // Outside of US Phone Number
        ],
    })
    const idsAndPhones = result.records.map(r => {
        const usPhone = r.getCellValueAsString("fldUB3ydfv356DV2c")
        const intlPhone = r.getCellValueAsString("fldvs0PPxIvoIWPzL")
        let canonicalPhone
        if (usPhone) {
            canonicalPhone = usPhoneIntoE164(usPhone)
        } else {
            canonicalPhone = intlPhoneIntoE164(intlPhone)
        }
        return {id: r.id, phone: canonicalPhone}
    })
    const records = idsAndPhones.filter(r => map.has(r.phone))
    output.text(`Found ${records.length} form records to be linked to the master table.`)
    for (let i = 0; i < records.length; i += 50) {
        output.text(`Processing updates in batch ${(i/50)+1}...`)
        const updates = []
        for (let j = i; j < records.length && j < 50; j++) {
            const record = records[j]
            const master = map.get(record.phone)
            let existingLinks = master.getCellValue(toFieldId)
            if (existingLinks) {
                if (existingLinks && existingLinks.map(v => v.id).includes(record.id)) {
                    // already linked, skip this update
                    continue
                }
                existingLinks.push({id: record.id})
            } else {
                existingLinks = [{id: record.id}]
            }
            const update = {
                id: master.id,
                fields: {[toFieldId]: existingLinks},
            }
            updates.push(update)
        }
        await masterTable.updateRecordsAsync(updates)
    }
    output.text(`Processed ${records.length} updates.`)
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
