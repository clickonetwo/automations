// noinspection SpellCheckingInspection

/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

import { base, remoteFetchAsync } from "airtable_internal";

const contactsTable = base.getTable("tblrsTMxY82X7DyZG"); // Contacts
const gbContactsTable = base.getTable("tbl7kftZWTbseOHis"); // GB Contacts

const contactsGbIdFieldId = "fldqqPq1b9b7VOVey"; // GiveButter ID
const contactsEmail1FieldId = "fldjGQciq7ccIRfsm"; // Email 1// Email 2
const contactsFirstNameFieldId = "fldlQ6XJ3xd1tQU1x"; // First Name
const contactsLastNameFieldId = "fld7BcJ3eYGn0xCWO"; // Last Name

const gbContactsGbIdFieldId = "fldG6oxvjWruz0UAv"; // ID

await doit();

async function doit() {
    const gbDonors = await fetchGbDonors();
    console.log(`Fetched ${Object.keys(gbDonors).length} donors from GiveButter`);
    const gbContacts = await fetchGbContacts();
    console.log(`Fetched ${gbContacts.length} known donors from GiveButter`);
    const contacts = await fetchAirtableContacts();
    console.log(`Fetched ${contacts.length} contacts with GiveButter IDs from Airtable`);
    // remove donors whose gbId is already on an Airtable contact
    for (const contact of contacts) {
        delete gbDonors[contact.gbId];
    }
    // remove donors whose gbId is already on a GB contact
    for (const gbContact of gbContacts) {
        delete gbDonors[gbContact.gbId];
    }
    // create a contact for each remaining gbDonor
    console.log(`Found ${Object.keys(gbDonors).length} unmatched donors from GiveButter`);
    const inserts = [];
    const listing = [];
    for (const payload of Object.values(gbDonors)) {
        listing.push({
            gbContactId: payload.contact_id.toString(),
            gbFirstName: payload.first_name,
            gbLastName: payload.last_name,
            gbEmail: payload.email,
            gbFirstDonationAmount: payload.donated,
            gbFirstDonationDate: payload.created_at,
            gbIsRecurring: payload.plan_id !== null,
        });
        inserts.push([
            { [contactsGbIdFieldId]: payload.contact_id.toString() },
            { [contactsEmail1FieldId]: payload.email },
            { [contactsFirstNameFieldId]: payload.first_name },
            { [contactsLastNameFieldId]: payload.last_name },
            { donated: payload.donated },
            { created_at: payload.created_at },
        ]);
    }
    console.log(JSON.stringify(listing, null, 2));
}

async function fetchGbDonors() {
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    let gbApiEndpoint = `https://api.givebutter.com/v1/transactions`;
    let page = 1;
    let more = true;
    /** @type {{[key: string]: GbDonationData}} */
    const donors = {};
    let donorCount = 0;
    while (more) {
        const response = await remoteFetchAsync(gbApiEndpoint + `?page=${page}`, {
            method: "GET",
            headers: {
                Accept: "application/json",
                Authorization: `Bearer ${gbApiKey}`,
            },
        });
        if (!response.ok) {
            throw new Error(
                `Fetch error on page ${page}: ${response.status}, ${response.statusText}`
            );
        }
        page++;
        /** @type {GbListDonationsPayload} */
        const body = await response.json();
        const last_page = body.meta.last_page ?? 0;
        more = last_page > page;
        for (const payload of body.data) {
            if (!payload.email || payload.donated < 1) {
                continue;
            }
            const donorId = payload.contact_id.toString();
            if (!donors[donorId]) {
                donors[donorId] = payload;
                donorCount++;
                if (donorCount % 100 === 0) {
                    console.log(`On page ${page} of ${last_page}, found ${donorCount} donors`);
                }
            }
        }
    }
    console.log(`After ${page} pages, found ${donorCount} donors with emails`);
    return donors;
}

async function fetchGbContacts() {
    const result = await gbContactsTable.selectRecordsAsync({
        fields: [gbContactsGbIdFieldId],
    });
    return result.records.map((r) => ({
        id: r.id,
        gbId: r.getCellValueAsString(gbContactsGbIdFieldId),
    }));
}

async function fetchAirtableContacts() {
    const result = await contactsTable.selectRecordsAsync({
        fields: [contactsGbIdFieldId],
    });
    return result.records
        .map((r) => ({
            id: r.id,
            gbId: r.getCellValueAsString(contactsGbIdFieldId),
        }))
        .filter((r) => r.gbId);
}
