/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

// Global definitions
const masterNamedFieldMap = {
    conflicts: "fldcohpR70JZIqqEl",             // Additional Form Submission Info
    asylumLinks: "fldTRnUG4ESOqteAg",           // Asylum Screening Form
    linkId: "fldBGO1zn0kcOYNEd",                // Asylum Screening Link ID
}

const formNamedFieldMap = {
    masterLink: "fld9j9o0qFtBczkjQ",            // Contact
    createdDate: "fldCqHWHyKZdLjZBQ",           // Created
}

const stringFieldMap = {  // map from field IDs in Asylum Screening Form table to All Contacts table
    "fldgrzsoSDC41zvUs": "fldG8MGdqTySJhHJ4",   // Preferred Name -> Preferred Name
    "fldlQlkYznc9z9uId": "flddEBS3pVsatT6Zh",   // A Number -> A Number
    "fldrCDTE7PYGibJOR": "fld1NoEE1EB3ckLFW",   // Questions/Additional Info -> Questions (Asylum Screening)
}

const dateFieldMap = {
    "fldQ56Cbf0pEW2cxx": "fld7HU3A7jWLbuzDU",   // Date of Most Recent Entry -> Date of Last Entry (Asylum Screening)
}

const singleChoiceFieldMap = {  // map from field IDs in Asylum Screening Form table to All Contacts table
    "fld7cZumyeYNTDWqt": "fldzxUADCkam6dLCR",   // How Entered US -> How Entered US (Asylum Screening)
    "fld3ZrmO7G6JhWtFm": "fldcuQYP0Fmb1SLwq",   // Current Immigration Status -> Current Immigration Status (Asylum Screening)
    "fldnd6G26IWb5hpx8": "fldzEWOEGf2hcD9Rt",   // Arrested in U.S.? -> Arrested in U.S. (Asylum Screening)
    "fldAMTgJzSF85YiIo": "fldwdXJ1MTAPuWHnz",   // Spoken to Lawyer? -> Spoken to Lawyer (Asylum Screening)
    "fldMptekewPll9U2a": "fld5g5l6x1qBli3HF",   // Spoken to Oasis before -> Prior Contact with Oasis (Asylum Screening)
    "fldIArklsIUvTcQJC": "fldXEd7mML9XRoBaj",   // Already Applied for Asylum? -> Already Filed I-589? (Asylum Screening)
}

const multiChoiceFieldMap = {    // map from field IDs in Asylum Screening Form table to All Contacts table
    "fldU3PLanlmM5yXDl": "fldRrH5d7z3uTxdQD",   // Pronouns -> Pronouns
}

const attachmentFieldMap = {  // map from field IDs in Asylum Screening Form table to All Contacts table
    "fldkIVOk69sSm4FAL": "fldtWTfL7aLfTjMLd",   // Documents -> Documents Provided
}

// Script invocation (as automation)
const formTable = base.getTable("tblEY7Yv4aiJxVSbf")    // Asylum Screening Form Table
const masterTable = base.getTable("tblsnJnJ4ubpZFLwq")  // All Contacts Master Table
const { formRecordId } = input.config()
const formRecord = await formTable.selectRecordAsync(formRecordId)
if (!formRecord) {
    throw new Error(`No asylum screening form record exists for ID ${formRecordId}`)
}
const { masterRecordId } = await extractMatchData()
await newAsylumScreeningRecordAction()

async function extractMatchData() {
    const masterLinks = formRecord.getCellValue(formNamedFieldMap.masterLink) || []
    if (!masterLinks.length) {
        throw new Error(`No master record ID in incoming asylum screening form ${formRecordId}!`)
    }
    const masterRecordId = masterLinks[0].id
    // console.log(JSON.stringify({ masterRecordId }, null, 2))
    return { masterRecordId }
}

async function newAsylumScreeningRecordAction() {
    // if this form refers to a specific record, update that one
    let masterRecord = await masterTable.selectRecordAsync(masterRecordId);
    if (!masterRecord) {
        throw new Error(`Form refers to master record ${masterRecordId}, but it doesn't exist`)
    }
    const hook = "https://hook.us1.make.com/h61316aga8y1rjfwkweqkz6rpjqtbo81"
    const linkId = masterRecord.getCellValue(masterNamedFieldMap.linkId)
    if (linkId) {
        await triggerLinkUpdate(hook, masterRecordId, linkId)
    } else {
        console.warn(`Form refers to master record ${masterRecordId}, but it has no link to update`)
    }
    await updateExistingMasterRecord(masterRecord)
}

async function updateExistingMasterRecord(masterRecord) {
    const multiForms = masterRecord.getCellValue(masterNamedFieldMap.asylumLinks)?.length !== 1
    const masterFields = {}
    let conflicts = ""
    for (let [srcKey, targetKey] of Object.entries(stringFieldMap)) {
        const master = masterRecord.getCellValueAsString(targetKey)
        const form = formRecord.getCellValueAsString(srcKey)
        if (!form || master === form) {
        } else if (!master) {
            masterFields[targetKey] = form
        } else {
            const fieldName = masterTable.getField(targetKey).name
            conflicts += `\t${fieldName}: ${form.replace(/\n/g,' ')}\n`
        }
    }
    for (let [srcKey, targetKey] of Object.entries(dateFieldMap)) {
        const master = masterRecord.getCellValue(targetKey) || ''
        const form = formRecord.getCellValue(srcKey) || ''
        if (!form || master === form) {
        } else if (!master) {
            masterFields[targetKey] = form
        } else {
            const fieldName = masterTable.getField(targetKey).name
            conflicts += `\t${fieldName}: ${form}\n`
        }
    }
    for (let [srcKey, targetKey] of Object.entries(singleChoiceFieldMap)) {
        const master = masterRecord.getCellValueAsString(targetKey)
        const form = formRecord.getCellValueAsString(srcKey)
        if (!form || master === form) {
        } else if (!master) {
            masterFields[targetKey] = { name: form }
        } else {
            const fieldName = masterTable.getField(targetKey).name
            conflicts += `\t${fieldName}: ${form}\n`
        }
    }
    for (let [srcKey, targetKey] of Object.entries(multiChoiceFieldMap)) {
        const master = masterRecord.getCellValue(targetKey) || []
        const masterVals = master.map(v => v.name).sort().join("|")
        const form = formRecord.getCellValue(srcKey) || []
        const formVals = form.map(v => v.name).sort().join("|")
        if (!formVals || masterVals === formVals) {
        } else if (!master.length) {
            masterFields[targetKey] = form.map(v => ({ name: v.name }))
        } else {
            const fieldName = masterTable.getField(targetKey).name
            const val = form.map(v => v.name).join(', ')
            conflicts += `\t${fieldName}: ${val}\n`
        }
    }
    for (let [srcKey, targetKey] of Object.entries(attachmentFieldMap)) {
        const master = masterRecord.getCellValue(targetKey) || []
        const form = formRecord.getCellValue(srcKey) || []
        if (form.length) {
            // console.log(`Sending attachments value: ${JSON.stringify(form, null, 2)}`)
            masterFields[targetKey] = [...master, ...form]
        }
    }
    if (conflicts) {
        const master = masterRecord.getCellValueAsString(masterNamedFieldMap.conflicts)
        const date = new Date(formRecord.getCellValue(formNamedFieldMap.createdDate))
        const timestamp = date.toLocaleString("en-US", { timeZone: "America/Los_Angeles" })
        const header = `Additional asylum screening data submitted ${timestamp}:`
        masterFields[masterNamedFieldMap.conflicts] = `${header}\n${conflicts}\n${master}`
    }
    if (Object.keys(masterFields).length) {
        await masterTable.updateRecordAsync(masterRecord.id, masterFields)
        if (multiForms && conflicts) {
            throw new Error(`Conflicting asylum screening forms for master record ${masterRecordId}`)
        }
    } else {
        console.warn(`Asylum Screening form ${formRecordId} had no additional information.`)
    }
}

async function markMasterRecordsAsDuplicates(masterTable, masterRecords) {
    const updates = masterRecords.map((r) => ({
        id: r.id,
        fields: { [masterNamedFieldMap.multiLink]: true },
    }))
    console.log(JSON.stringify(updates, null, 2))
    for (let i = 0; i < updates.length; i += 50) {
        const end = Math.min(updates.length, i + 50)
        await masterTable.updateRecordsAsync(updates.slice(i, end))
    }
}

async function triggerLinkUpdate(hook, recordId, linkId) {
    console.log(`Requesting reset of link '${linkId}' from record ${recordId}...`);
    const response = await fetch(hook, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            recordId: recordId,
            shortLinkId: linkId
        })
    }).catch((error) => {
        console.error("Error sending request to make:", error);
    })
    if (response) {
        console.log(`Response code from make is ${response.status}`);
    }
}
