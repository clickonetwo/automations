// noinspection SpellCheckingInspection

/*
 * Copyright 2024-2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

import {base, input} from "airtable_internal";

const donationsTable = base.getTable('tblAjtXdKANH06DIn')
const contactsTable = base.getTable('tbl7kftZWTbseOHis')
const plansTable = base.getTable('tblqgeY56ThyqHjIM')

const donationsPayloadFieldId = 'fldxlf48Q1rsQ41XM'
const donationsIdFieldId = 'fldMw3mEFuRy5dAG4'
const donationsDateFieldId = 'fldEcuCszWu573yiU'
const donationsAmountFieldId = 'fldbA1gIdCQ5LHWu1'

const contactsIdFieldId = 'fldG6oxvjWruz0UAv'
const contactsFirstNameFieldId = 'fldgUmNqkmBkbi5gg'
const contactsLastNameFieldId = 'flduCty6t3vCBs0hw'
const contactsEmailFieldId = 'fldNeK7kFqJLj2cWj'
const contactsDonationsFieldId = 'fld3Td5bZlyhDVktH'

const plansIdFieldId = 'fld345XhcHaQV5Oaj'
const plansDonorsFieldId = 'fld4fzS8m4hQjBvT0'
const plansDonationsFieldId = 'fldhdn7ccLa51a0bl'
const plansStatusFieldId = 'fld70U96Fh0erAd27'
const plansFrequencyFieldId = 'fldPF06oq9X6h1MMn'
const plansStartedFieldId = 'fldxDW7rQGrV9iQGd'
const plansEndedFieldId = 'fldZk7rhKQxmBIxin'

const { donationRecordId } = input.config()
const donationRecord = await donationsTable.selectRecordAsync(donationRecordId)
if (!donationRecord) {
    throw new Error(`No donation record exists for ID ${donationRecordId}`)
}
await processDonationRecord()

async function processDonationRecord() {
    const payload = JSON.parse(donationRecord.getCellValue(donationsPayloadFieldId)).data
    if (!payload) {
        throw new Error(`No payload data for donation record ${donationRecordId}`)
    }
    // see if we already have this donation event
    const existing = await donationsTable.selectRecordsAsync({
        fields: [donationsIdFieldId]
    })
    const matchingRecords = existing.records.filter(r => r.getCellValue(donationsIdFieldId) === payload.id)
    if (matchingRecords.length) {
        if (matchingRecords.length > 1) {
            throw new Error(`There are multiple donation records with id ${payload.id}`)
        } else if (matchingRecords[0].id !== donationRecordId) {
            throw new Error(`This payload was already processed in donation record ${matchingRecords[0].id}`)
        } else {
            console.warn(`Donation record ${donationRecordId} is being reprocessed`)
        }
    }
    const fieldUpdates = {
        [donationsIdFieldId]: payload.id,
        [donationsDateFieldId]: new Date(payload.created_at).toISOString(),
    }
    if (payload.donated >= 0) {
        if (payload.donated === 0) {
            console.warn(`Donation record ${donationRecordId} has a zero amount`)
        }
        fieldUpdates[donationsAmountFieldId] = payload.donated
    } else {
        throw new Error(`Received donation record ${donationRecordId} with negative amount ${payload.donated}`)
    }
    await donationsTable.updateRecordAsync(donationRecordId, fieldUpdates)
    const donorRecordId = await createOrUpdateDonor(payload)
    if (payload.plan_id) {
        await createOrUpdatePlan(payload.plan_id, donorRecordId)
    }
}

async function createOrUpdateDonor(payload) {
    const donorId = payload.contact_id.toString()
    const existing = await contactsTable.selectRecordsAsync({
        fields: [contactsIdFieldId, contactsDonationsFieldId]
    })
    const matchingRecords = existing.records.filter(r => r.getCellValue(contactsIdFieldId) === donorId)
    // if there is a matching contact, make sure this donation is linked
    if (matchingRecords.length) {
        if (matchingRecords.length > 1) {
            throw new Error(`There is more than one donor with ID ${donorId}`)
        }
        const donorRecord = matchingRecords[0]
        const existing = donorRecord.getCellValue(contactsDonationsFieldId)
        if (existing && !existing.map(v => v.id).includes(donationRecordId)) {
            console.log(`This donation link is being added to existing contact ID ${donorId}`)
            await contactsTable.updateRecordAsync(donorRecord.id, {
                [contactsDonationsFieldId]: [ ...existing, { id: donationRecordId } ]
            })
        } else {
            console.log(`This donation is already linked to existing contact ID ${donorId}`)
        }
        return donorRecord.id
    }
    // if there is no matching contact, create one and link this donation
    console.log(`Contact ${donorId} is being added and linked to this donation`)
    const newFields = {
        [contactsIdFieldId]: donorId,
        [contactsFirstNameFieldId]: payload.first_name,
        [contactsLastNameFieldId]: payload.last_name,
        [contactsEmailFieldId]: payload.email,
        [contactsDonationsFieldId]: [{ id: donationRecordId }],
    }
    return await contactsTable.createRecordAsync(newFields)
}

async function createOrUpdatePlan(planId, donorRecordId) {
    const existing = await plansTable.selectRecordsAsync({
        fields: [plansIdFieldId, plansDonorsFieldId, plansDonationsFieldId]
    })
    const matchingRecords = existing.records.filter(r => r.getCellValue(plansIdFieldId) === planId)
    // if there is a matching plan, make sure this donation is linked
    if (matchingRecords.length) {
        if (matchingRecords.length > 1) {
            throw new Error(`There is more than one plan with ID ${planId}`)
        }
        const planRecord = matchingRecords[0]
        let planUpdates = []
        const existingDonors = planRecord.getCellValue(plansDonorsFieldId)
        if (existingDonors && !existingDonors.map(v => v.id).includes(donorRecordId)) {
            console.log(`Donor is being added to existing plan ${planId}`)
            planUpdates.push([ plansDonorsFieldId, [...existingDonors, {id: donorRecordId}] ])
        } else {
            console.log(`Donor is already linked to existing plan ${planId}`)
        }
        const existingDonations = planRecord.getCellValue(plansDonationsFieldId)
        if (existingDonations && !existingDonations.map(v => v.id).includes(donationRecordId)) {
            console.log(`Donation is being added to existing plan ${planId}`)
            planUpdates.push([ plansDonationsFieldId, [ ...existingDonations, { id: donationRecordId }] ])
        } else {
            console.log(`Donation is already linked to existing plan ${planId}`)
        }
        if (planUpdates.length) {
            await plansTable.updateRecordAsync(planRecord.id, Object.fromEntries(planUpdates))
        }
        return
    }
    // otherwise fetch and create the matching plan
    console.log(`Plan ${planId} is being fetched and linked to this donation`)
    const payload = await fetchPlan(planId)
    const fieldValues = {
        [plansIdFieldId]: planId,
        [plansDonorsFieldId]: [ { id: donorRecordId } ],
        [plansDonationsFieldId]: [{ id: donationRecordId } ],
        [plansStatusFieldId]: payload.status,
        [plansFrequencyFieldId]: payload.frequency,
        [plansStartedFieldId]: payload.start_at.slice(0, 10),
    }
    if (payload.canceled_at) {
        fieldValues[plansEndedFieldId] = payload.canceled_at.slice(0, 10)
    }
    await plansTable.createRecordAsync(fieldValues)
}

async function fetchPlan(planId) {
    const gbApiKey = '8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs'
    const gbApiEndpoint = `https://api.givebutter.com/v1/plans/${planId}`
    const response = await fetch(gbApiEndpoint, {
        method: 'GET',
        headers: {
            Accept: 'application/json',
            Authorization: `Bearer ${gbApiKey}`
        }
    })
    if (!response.ok) {
        throw new Error(JSON.stringify(response))
    }
    return await response.json()
}
