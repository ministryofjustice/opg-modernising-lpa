import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('donor cannot sign LPA', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Add a correspondent'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Provide your details'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Can you sign the LPA yourself online?');
    await page.getByRole('radio', {name: 'No', exact: true}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Check that you can sign your LPA');
    await page.getByRole('radio', {name: 'No'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.getByLabel('Information saved').getByRole('paragraph')).toContainText('We know that you will need help signing your LPA.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Return to task list'}).click();
    await page.getByRole('link', {name: 'Choose your signatory and'}).click();

    await expect(page.locator('h1')).toContainText('Getting help signing your LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Enter the name of your authorised signatory');
    await page.getByRole('textbox', {name: 'First names'}).fill('Davie');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Jones');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Enter the name of your independent witness');
    await page.getByRole('textbox', {name: 'First names'}).fill('Indie');
    await page.getByRole('textbox', {name: 'Last name'}).fill('White');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Enter the mobile number of your independent witness');
    await page.getByRole('textbox', {name: 'UK mobile number'}).fill('07700900000');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Add Indie White’s address');
    await page.getByRole('radio', {name: 'Use an address you’ve already'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: '2 RICHMOND PLACE KINGS HEATH'}).check();

    await expect(page.locator('h1')).toContainText('Select an address for Indie White');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Check and send to your'}).click();

    await expect(page.locator('h1')).toContainText('Check your LPA');
    await page.getByRole('checkbox', {name: 'I’ve checked this LPA and I’m'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Confirm'}).click();
    await page.getByRole('link', {name: 'Return to task list'}).click();
    await page.getByRole('link', {name: 'Pay for the LPA'}).click();

    await expect(page.locator('h1')).toContainText('Paying for your LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'No'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByText('Card number Accepted credit').click();
    await page.getByRole('textbox', {name: 'Card number'}).fill('4444333322221111');
    await page.getByRole('textbox', {name: 'Month'}).fill('10');
    await page.getByRole('textbox', {name: 'Year'}).fill('27');
    await page.getByRole('textbox', {name: 'Name on card'}).fill('Name Card');
    await page.getByRole('textbox', {name: 'Card security code'}).fill('123');
    await page.getByRole('textbox', {name: 'Address line 1'}).fill('1 Richmond');
    await page.getByRole('textbox', {name: 'Town or city'}).fill('Birmingham');
    await page.getByRole('textbox', {name: 'Postcode'}).fill('B14 7ED');
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('button', {name: 'Confirm payment'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Confirm your identity'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Return to task list'}).click();
    await page.getByRole('link', {name: 'Sign the LPA'}).click();

    await expect(page.locator('h1')).toContainText('How to sign your LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Start'}).click();

    await expect(page.locator('h1')).toContainText('Your LPA will be registered in English');
    await page.getByRole('radio', {name: 'Register my LPA in English'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Read your LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Your legal rights and responsibilities');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue to signing page'}).click();

    await expect(page.locator('h1')).toContainText('Sign your LPA');
    await page.getByRole('checkbox', {name: 'Sam Smith wants to sign this'}).check();
    await page.getByRole('checkbox', {name: 'Sam Smith wants to apply to'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Submit signature'}).click();

    await expect(page.locator('h1')).toContainText('Witnessing your signature');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Indie White, confirm you witnessed the donor sign their LPA');
    await page.getByRole('textbox', {name: 'Enter code'}).fill('1234');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Charlie Cooper, confirm you witnessed the donor sign their LPA');
    await page.getByRole('textbox', {name: 'Enter code'}).fill('1234');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('You’ve submitted your LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();
});
