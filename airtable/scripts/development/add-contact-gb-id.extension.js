// noinspection SpellCheckingInspection

/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

const contactsTable = base.getTable('tblrsTMxY82X7DyZG')

const contactsGbIdFieldId = 'fldqqPq1b9b7VOVey'     // GiveButter ID
const contactsEmail1FieldId = 'fldjGQciq7ccIRfsm'   // Email 1
const contactsEmail2FieldId = 'fldGHRReWTPk49t1Y'   // Email 2

await doit()

async function doit() {
    const gbContacts = await fetchGbContacts()
    console.log(`Fetched ${gbContacts.length} contacts with emails from GiveButter`)
    const contacts = await fetchAirtableContacts()
    console.log(`Fetched ${contacts.length} contacts with emails from Airtable`)
    const updates = []
    let matches = 0
    for (const contact of contacts) {
        for (const gbContact of gbContacts) {
            if (gbContact.email === contact.email1 || gbContact.email === contact.email2) {
                matches++
                updates.push({id: contact.id, fields: {[contactsGbIdFieldId]: gbContact.id}})
                break
            }
        }
        if (matches+1 % 100 === 0) console.log(
            `Matched ${matches} contact emails with contact emails from GiveButter`
        )
    }
    console.log(`Found ${updates.length} contacts to update`)
    for (let i = 0; i < updates.length; i += 50) {
        const end = Math.min(updates.length, i + 50)
        await contactsTable.updateRecordsAsync(updates.slice(i, end))
    }
}


async function fetchGbContacts() {
    /** @type {{id: string, email: string}[]} */
    const contacts = []
    const gbApiKey = '8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs'
    let gbApiEndpoint = `https://api.givebutter.com/v1/contacts`
    let page = 0
    let total = 0
    while (gbApiEndpoint) {
        const response = await remoteFetchAsync(gbApiEndpoint, {
            method: 'GET',
            headers: {
                Accept: 'application/json',
                Authorization: `Bearer ${gbApiKey}`
            }
        })
        if (!response.ok) {
            throw new Error(JSON.stringify(response))
        }
        page++
        /** @type {GbListContactsPayload} */
        const body = await response.json()
        gbApiEndpoint = body.links.next
        for (const contact of body.data) {
            if (contact.primary_email) {
                contacts.push({id: contact.id.toString(), email: contact.primary_email.toLowerCase()})
                total++
                if (total % 100 === 0) {
                    console.log(`As of page ${page}, found ${total} contacts with emails`)
                }
            }
        }
    }
    console.log(`After ${page} pages, found ${total} contacts with emails`)
    return contacts
}

async function fetchAirtableContacts() {
    const result = await contactsTable.selectRecordsAsync({
        fields: [contactsEmail1FieldId, contactsEmail2FieldId]
    })
    let table = result.records.map(r => ({
        id: r.id,
        email1: r.getCellValueAsString(contactsEmail1FieldId).toLowerCase(),
        email2: r.getCellValueAsString(contactsEmail2FieldId).toLowerCase(),
    }))
    table = table.filter(r => r.email1 || r.email2)
    return table
}
