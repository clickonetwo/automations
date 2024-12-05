await linkAsylumQuestionnaireExtension()

async function linkAsylumQuestionnaireExtension() {
    const masterTable = base.getTable("tblsnJnJ4ubpZFLwq")
    output.text(`Finding records with questionnaires to be linked...`)
    const found = await findRecordsWithQuestionnaireIds(masterTable)
    output.text(`Found ${found.length} records to link.`)
    await linkRecordsWithQuestionnaireIds(masterTable, found)
}

async function findRecordsWithQuestionnaireIds(masterTable) {
    let result = await masterTable.selectRecordsAsync({
        fields: ["fldHcio0hgdSJZFcn"],    // Asylum Screening Record ID for Migration
    })
    return result.records.filter(r => (r.getCellValueAsString("fldHcio0hgdSJZFcn") !== ""))
}

async function linkRecordsWithQuestionnaireIds(masterTable, records) {
    for (let i = 0; i < records.length; i += 50) {
        output.text(`Processing relink of 50 records in batch ${i+1}...`)
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
