import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('donor has not known certificate provider long enough to carry out role', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Add restrictions to the LPA'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Choose your certificate'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'First names'}).fill('Charlie');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Cooper');
    await page.getByRole('textbox', {name: 'UK mobile number'}).click();
    await page.getByRole('textbox', {name: 'UK mobile number'}).fill('07700900000');
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Personally'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('How long have you known Charlie Cooper?');
    await page.getByRole('radio', {name: 'Less than 2 years'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('You must choose a different certificate provider');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
});
