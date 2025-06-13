import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";


test('donor applies for a fee remission and chooses to send evidence in the post', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Check and send to your'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Pay for the LPA'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Half fee (a remission)'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Evidence required to pay a half fee');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Send it by post'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Send us your evidence by post');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue to payment'}).click();

    await expect(page.getByLabel('Important')).toContainText('This is a test page. No money will be taken.');
    await page.getByRole('textbox', {name: 'Card number'}).click();
    await page.getByRole('textbox', {name: 'Card number'}).fill('4444333322221111');
    await page.getByRole('textbox', {name: 'Month'}).fill('10');
    await page.getByRole('textbox', {name: 'Year'}).fill('27');
    await page.getByRole('textbox', {name: 'Name on card'}).fill('Name Card');
    await page.getByRole('textbox', {name: 'Card security code'}).fill('123');
    await page.getByRole('textbox', {name: 'Address line 1'}).fill('1');
    await page.getByRole('textbox', {name: 'Enter address line'}).fill('Richmond Place');
    await page.getByRole('textbox', {name: 'Postcode'}).fill('B14 7ED');
    await page.getByRole('textbox', {name: 'Town or city'}).fill('Birmingham');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Confirm your payment');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Confirm payment'}).click();

    await expect(page.locator('h1')).toContainText('Payment received');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.getByLabel('Important').getByRole('paragraph')).toContainText('We’ll review the evidence you send about your LPA fee');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Return to task list'}).click();
});
