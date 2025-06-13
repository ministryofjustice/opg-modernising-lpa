import {expect, test} from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";
import path from 'path.js'

test('donor makes a repeat application and provides a legacy (7) reference number', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Check and send to your'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Pay for the LPA'}).click();
    await page.getByText('Continue Return to task list').click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Repeat application discount'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('What is your previous reference number?');
    await page.getByRole('textbox', {name: 'Previous reference number'}).click();
    await page.getByRole('textbox', {name: 'Previous reference number'}).fill('7000 0000 0000');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('How much did you previously pay for your LPA?');
    await page.getByRole('radio', {name: '£41 (half fee)'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('New evidence required to pay a half fee');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('How would you like to send us your evidence?');
    await page.getByRole('radio', {name: 'Upload it online'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('button', {name: 'Upload a file'}).setInputFiles([
        path.join(__dirname, 'upload-file-1.jpg'),
        path.join(__dirname, 'upload-file-2.jpg'),
    ]);
    await page.getByRole('button', {name: 'Upload files'}).click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue to payment'}).click();

    await expect(page.getByLabel('Important')).toContainText('This is a test page. No money will be taken.');
    await page.getByRole('textbox', {name: 'Card number'}).fill('4444333322221111');
    await page.getByRole('textbox', {name: 'Month'}).fill('10');
    await page.getByRole('textbox', {name: 'Year'}).fill('27');
    await page.getByRole('textbox', {name: 'Name on card'}).fill('Name Card');
    await page.getByRole('textbox', {name: 'Card security code'}).fill('123');
    await page.getByRole('textbox', {name: 'Address line 1'}).fill('1 Richmond Place');
    await page.getByRole('textbox', {name: 'Town or city'}).fill('Birmingham');
    await page.getByRole('textbox', {name: 'Postcode'}).fill('B14 7ED');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.getByLabel('Important')).toContainText('This is a test page. No money will be taken.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Confirm payment'}).click();

    await expect(page.locator('h1')).toContainText('Payment received');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Evidence successfully uploaded');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Return to task list'}).click();
});
