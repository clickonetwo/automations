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
    output.text(`Found ${found.length} records to link.`)
    return result.records.filter(r => (r.getCellValueAsString("fldHcio0hgdSJZFcn") !== ""))
}

async function linkRecordsWithQuestionnaireIds(masterTable, records) {
    for (let i = 0; i < records.length; i += 50) {
        output.text(`Processing links in batch ${(i/50)+1}...`)
        const updates = []
        for (let j = i; j < records.length && j < 50; j++) {
            const record = records[j]
            const linkId = record.getCellValueAsString("fldHcio0hgdSJZFcn")
            const update = {
                id: record.id,
                fields: {"fldmk8m8u8nojUffm": [{id: linkId}]},
            }
            updates.push(update)
        }
        await masterTable.updateRecordsAsync(updates)
    }
    output.text(`Processed ${records.length} records.`)
}
