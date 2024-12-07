/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

await linkAsylumQuestionnaireExtension()

async function linkAsylumQuestionnaireExtension() {
    const masterTable = base.getTable("tblsnJnJ4ubpZFLwq")
    const found = await findRecordsWithQuestionnaireIds(masterTable)
    await linkRecordsWithQuestionnaireIds(masterTable, found)
}

async function findRecordsWithQuestionnaireIds(masterTable) {
    output.text(`Finding records with questionnaires to be linked...`)
    let result = await masterTable.selectRecordsAsync({
        fields: ["fldHcio0hgdSJZFcn"],    // Asylum Screening Record ID for Migration
    })
    let records = result.records.map(r => {
        const links = r.getCellValueAsString("fldHcio0hgdSJZFcn")
        if (!links) {
            return {id: r.id, records: []}
        }
        const linksArray = links.split(", ")
        return {id: r.id, records: linksArray}
    })
    records = records.filter(r => (r.records.length > 0))
    output.text(`Found ${records.length} records to link.`)
    return records
}

async function linkRecordsWithQuestionnaireIds(masterTable, records) {
    for (let i = 0; i < records.length; i += 50) {
        output.text(`Processing links in batch ${(i/50)+1}...`)
        const updates = []
        for (let j = i; j < records.length && j < 50; j++) {
            const record = records[j]
            const links = record.records.map(rId => ({id: rId}))
            const update = {
                id: record.id,
                fields: {"fldmk8m8u8nojUffm": links},
            }
            updates.push(update)
        }
        await masterTable.updateRecordsAsync(updates)
    }
    output.text(`Processed ${records.length} records.`)
}
