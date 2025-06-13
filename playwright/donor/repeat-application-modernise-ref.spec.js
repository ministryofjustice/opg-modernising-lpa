import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";


test('donor makes a repeat application and provides a modernised (M) reference number', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Check and send to your'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Pay for the LPA'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Repeat application discount'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('textbox', {name: 'Previous reference number'}).click();
    await page.getByRole('textbox', {name: 'Previous reference number'}).fill('M123 4567 1234');
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('The cost of your repeat application');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('radio', {name: 'OPG has told me I am eligible to pay no fee'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.getByLabel('Success')).toContainText('Repeat application no fee request submitted.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Return to task list'}).click();
});
