import {expect, test} from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('donor triggers post office ID route and then chooses a voucher', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Pay for the LPA'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Confirm your identity'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.goBack()
    await page.goBack()
    await page.getByRole('link', {name: 'Confirm your identity'}).click();

    await expect(page.locator('#main-content')).toContainText('We need more information about how you will confirm your identity');
    await page.getByRole('radio', {name: 'I want to return to GOV.UK'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Unable to prove identity (X)'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.getByLabel('Important').getByRole('paragraph')).toContainText('You were not able to confirm your identity using GOV.UK One Login.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Choose someone to vouch for you');
    await page.getByRole('radio', {name: 'Yes, I know someone who is'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Enter details for the person confirming your identity');
    await page.getByRole('textbox', {name: 'First names'}).fill('Vivian');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Smith');
    await page.getByRole('textbox', {name: 'Email address'}).fill('simulate-delivered@notifications.service.gov.uk');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('#main-content')).toContainText('Are you sure this person can confirm your identity?');
    await page.getByRole('radio', {name: 'Yes'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Check your details');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('We have contacted Vivian Smith to confirm your identity');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
