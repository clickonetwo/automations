// noinspection SpellCheckingInspection

/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

import { base, remoteFetchAsync } from "airtable_internal";

const contactsTable = base.getTable("tblrsTMxY82X7DyZG"); // Contacts
const donorsTable = base.getTable("tbl7kftZWTbseOHis"); // GB Contacts

const contactsGbIdFieldId = "fldqqPq1b9b7VOVey"; // GiveButter ID
const contactsEmail1FieldId = "fldjGQciq7ccIRfsm"; // Email 1
const contactsEmail2FieldId = "fldGHRReWTPk49t1Y"; // Email 2

const donorsEmailFieldId = "fldNeK7kFqJLj2cWj";
const donorsNewContactFieldId = "fld7boBE7WgjzOAhD";

await doit();

async function doit() {
    const gbContacts = await fetchGbContacts();
    console.log(`Fetched ${gbContacts.length} contacts with emails from GiveButter`);
    const gbDonors = await fetchGbDonors();
    console.log(`Fetched ${gbDonors.length} donors with emails from GiveButter`);
    const contacts = await fetchAirtableContacts();
    console.log(`Fetched ${contacts.length} contacts with emails from Airtable`);
    const contactUpdates = [];
    const donorUpdates = [];
    const unmatched = [];
    gbloop: for (const gbContact of gbContacts) {
        for (const contact of contacts) {
            if (gbContact.email === contact.email1 || gbContact.email === contact.email2) {
                contactUpdates.push({
                    id: contact.id,
                    fields: { [contactsGbIdFieldId]: gbContact.id },
                });
                continue gbloop;
            }
        }
        for (const gbDonor of gbDonors) {
            if (gbContact.email === gbDonor.email) {
                donorUpdates.push({
                    id: gbDonor.id,
                    fields: { [donorsNewContactFieldId]: true },
                });
                continue gbloop;
            }
        }
        // unmatched email
        unmatched.push(gbContact);
    }
    console.log(`Found ${contactUpdates.length} contacts to update`);
    for (let i = 0; i < contactUpdates.length; i += 50) {
        const end = Math.min(contactUpdates.length, i + 50);
        await contactsTable.updateRecordsAsync(contactUpdates.slice(i, end));
    }
    console.log(`Found ${donorUpdates.length} donors to update`);
    for (let i = 0; i < donorUpdates.length; i += 50) {
        const end = Math.min(donorUpdates.length, i + 50);
        await donorsTable.updateRecordsAsync(donorUpdates.slice(i, end));
    }
    console.log(
        `Found ${unmatched.length} unmatched emails of GiveButter contacts:`,
        JSON.stringify(unmatched, null, 2)
    );
}

async function fetchGbContacts() {
    /** @type {{id: string, email: string}[]} */
    const contacts = [];
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    let gbApiEndpoint = `https://api.givebutter.com/v1/contacts`;
    let page = 0;
    let total = 0;
    while (gbApiEndpoint) {
        const response = await remoteFetchAsync(gbApiEndpoint, {
            method: "GET",
            headers: {
                Accept: "application/json",
                Authorization: `Bearer ${gbApiKey}`,
            },
        });
        if (!response.ok) {
            throw new Error(JSON.stringify(response));
        }
        page++;
        /** @type {GbListContactsPayload} */
        const body = await response.json();
        gbApiEndpoint = body.links.next;
        for (const contact of body.data) {
            if (contact.primary_email) {
                contacts.push({
                    id: contact.id.toString(),
                    email: contact.primary_email.toLowerCase(),
                    firstName: contact.first_name,
                    lastName: contact.last_name,
                });
                total++;
                if (total % 100 === 0) {
                    console.log(`As of page ${page}, found ${total} contacts with emails`);
                }
            }
        }
    }
    console.log(`After ${page} pages, found ${total} contacts with emails`);
    return contacts;
}

async function fetchGbDonors() {
    const result = await donorsTable.selectRecordsAsync({
        fields: [donorsEmailFieldId],
    });
    return result.records
        .map((r) => ({
            id: r.id,
            email: r.getCellValueAsString(donorsEmailFieldId).toLowerCase(),
        }))
        .filter((r) => r.email);
}

async function fetchAirtableContacts() {
    const result = await contactsTable.selectRecordsAsync({
        fields: [contactsEmail1FieldId, contactsEmail2FieldId],
    });
    let table = result.records.map((r) => ({
        id: r.id,
        email1: r.getCellValueAsString(contactsEmail1FieldId).toLowerCase(),
        email2: r.getCellValueAsString(contactsEmail2FieldId).toLowerCase(),
    }));
    table = table.filter((r) => r.email1 || r.email2);
    return table;
}
