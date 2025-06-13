

import { test, expect } from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('donor property and affairs full journey', async ({ page }) => {
  await page.goto('/start');

  await expect(page.locator('#main-content')).toContainText('Register a lasting power of attorney');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: /Start|Start now/ }).click();
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Manage your LPAs');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: /Start|Start now/ }).click();

  await expect(page.locator('h1')).toContainText('What is your name?');
  await page.getByRole('textbox', { name: 'First names' }).fill('Sam');
  await page.getByRole('textbox', { name: 'Last name' }).fill('Smith');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('What is your date of birth?');
  await page.getByRole('textbox', { name: 'Day' }).fill('02');
  await page.getByRole('textbox', { name: 'Month' }).fill('01');
  await page.getByRole('textbox', { name: 'Year' }).fill('2000');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Do you live in the UK, the Channel Islands or the Isle of Man?');
  await page.getByRole('radio', { name: 'Yes' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Enter your postcode');
  await page.getByRole('textbox', { name: 'Postcode' }).fill('B14 7ED');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Find address' }).click();

  await expect(page.locator('h1')).toContainText('Select your address');
  await page.getByLabel('Select an address').selectOption('{"line1":"1 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Confirm your address');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Receiving updates about your LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Can you sign the LPA yourself online?');
  await page.getByRole('radio', { name: 'Yes' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Your preferred language');
  await page.getByRole('group', { name: 'Which language would you prefer us to use when we contact you?' }).getByLabel('English').check();
  await page.getByRole('group', { name: 'In which language would you' }).getByLabel('English').check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Your legal rights and responsibilities if you make an LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Choose which type of LPA you want to make');
  await page.getByRole('radio', { name: 'Property and affairs LPA' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Your task list');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Choose your attorneys' }).click();

  await expect(page.locator('h1')).toContainText('Choosing your attorneys');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();
  await page.getByRole('textbox', { name: 'First names' }).fill('Jessie');
  await page.getByRole('textbox', { name: 'Last name' }).fill('Jones');
  await page.getByRole('textbox', { name: 'Day' }).fill('02');
  await page.getByRole('textbox', { name: 'Month' }).fill('01');
  await page.getByRole('textbox', { name: 'Year' }).fill('2000');
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Add Jessie Jones’ address');
  await page.getByRole('radio', { name: 'Enter a new address' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('textbox', { name: 'Postcode' }).fill('B14 7ED');

  await expect(page.locator('h1')).toContainText('What is Jessie Jones’ postcode?');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Find address' }).click();

  await expect(page.locator('h1')).toContainText('Select an address for Jessie Jones');
  await page.getByLabel('Select an address').selectOption('{"line1":"2 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Jessie Jones’ address');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('You have added 1 attorney');
  await page.getByRole('radio', { name: 'Yes' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('textbox', { name: 'First names' }).fill('Robin');
  await page.getByRole('textbox', { name: 'Last name' }).fill('Redcar');
  await page.getByRole('textbox', { name: 'Day' }).fill('02');
  await page.getByRole('textbox', { name: 'Month' }).fill('01');
  await page.getByRole('textbox', { name: 'Year' }).fill('2000');
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByText('Enter a new address').click();
  await page.getByRole('radio', { name: 'Use an address you’ve already' }).check();
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('radio', { name: '2 RICHMOND PLACE BIRMINGHAM' }).check();

  await expect(page.locator('h1')).toContainText('Select an address for Robin Redcar');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('radio', { name: 'No' }).check();
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('How your attorneys should make decisions');
  await page.getByText('Jointly and severally', { exact: true }).click();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('link', { name: 'Choose your replacement' }).click();

  await expect(page.locator('h1')).toContainText('Do you want any replacement attorneys?');
  await page.getByRole('radio', { name: 'Yes, I want replacement' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Who would you like to be your replacement attorney?');
  await page.getByRole('textbox', { name: 'First names' }).fill('Blake');
  await page.getByRole('textbox', { name: 'Last name' }).fill('Buckley');
  await page.getByRole('textbox', { name: 'Day' }).fill('02');
  await page.getByRole('textbox', { name: 'Month' }).fill('01');
  await page.getByRole('textbox', { name: 'Year' }).fill('2000');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('radio', { name: 'Use an address you’ve already' }).check();
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('radio', { name: '2 RICHMOND PLACE BIRMINGHAM' }).check();
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('radio', { name: 'Yes' }).check();
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('textbox', { name: 'First names' }).fill('Taylor');
  await page.getByRole('textbox', { name: 'Last name' }).fill('Thompson');
  await page.getByRole('textbox', { name: 'Day' }).fill('02');
  await page.getByRole('textbox', { name: 'Month' }).fill('01');
  await page.getByRole('textbox', { name: 'Year' }).fill('2000');
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('radio', { name: 'Use an address you’ve already' }).check();
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('radio', { name: '2 RICHMOND PLACE BIRMINGHAM' }).check();
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('radio', { name: 'No' }).check();
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('When your replacement attorneys step in');
  await page.getByRole('radio', { name: 'All together, as soon as one' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('link', { name: 'Choose when your LPA can be' }).click();

  await expect(page.locator('h1')).toContainText('When your attorneys can use your LPA');
  await page.getByRole('radio', { name: 'Whether or not I have mental' }).check()
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('link', { name: 'Add restrictions to the LPA' }).click();

  await expect(page.locator('h1')).toContainText('Restrict the decisions your attorneys can make');
  await page.getByRole('textbox', { name: 'Restrictions and conditions (' }).fill('My attorneys must not sell my home unless I need to fund a care home place.');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('link', { name: 'Choose your certificate' }).click();

  await expect(page.locator('h1')).toContainText('What a certificate provider does');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Choose your certificate provider');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Your certificate provider’s details');
  await page.getByRole('textbox', { name: 'First names' }).fill('Charlie');
  await page.getByRole('textbox', { name: 'Last name' }).fill('Cooper');
  await page.getByRole('textbox', { name: 'UK mobile number' }).fill('07700900000');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('How do you know Charlie Cooper, your certificate provider?');
  await page.getByRole('radio', { name: 'Personally' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('How long have you known Charlie Cooper?');
  await page.getByText('years or more').click();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('How would Charlie like us to contact them?');
  await page.getByRole('radio', { name: 'By email' }).check();
  await page.getByRole('textbox', { name: 'Certificate provider’s email' }).fill('simulate-delivered@notifications.service.gov.uk');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Add Charlie Cooper’s address');
  await page.getByRole('radio', { name: 'Enter a new address' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('textbox', { name: 'Postcode' }).fill('B14 7ED');
  await page.getByRole('button', { name: 'Find address' }).click();

  await expect(page.locator('h1')).toContainText('Select an address for Charlie Cooper');
  await page.getByLabel('Select an address').selectOption('{"line1":"5 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Charlie Cooper’s address');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('You’ve added a certificate provider');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Return to task list' }).click();
  await page.getByRole('link', { name: 'People to notify about your' }).click();

  await expect(page.locator('h1')).toContainText('People to notify about your LPA');
  await page.getByRole('radio', { name: 'No' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('link', { name: 'People to notify about your' }).click();
  await page.getByRole('radio', { name: 'Yes' }).check();
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Add a person to notify about your LPA');
  await page.getByRole('textbox', { name: 'First names' }).fill('Jordan');
  await page.getByRole('textbox', { name: 'Last name' }).fill('Jefferson');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Add Jordan Jefferson’s address');
  await page.getByRole('radio', { name: 'Enter a new address' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('textbox', { name: 'Postcode' }).fill('B14 7ED');
  await page.getByRole('button', { name: 'Find address' }).click();

  await expect(page.locator('h1')).toContainText('Select an address for Jordan Jefferson');
  await page.getByLabel('Select an address').selectOption('{"line1":"4 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Jordan Jefferson’s address');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('People to notify about your LPA');
  await page.getByRole('radio', { name: 'No' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('link', { name: 'Add a correspondent' }).click();

  await expect(page.locator('h1')).toContainText('Add a correspondent');
  await page.getByRole('radio', { name: 'No' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('link', { name: 'Check and send to your' }).click();

  await expect(page.locator('h1')).toContainText('Check your LPA');
  await page.getByRole('checkbox', { name: 'I’ve checked this LPA and I’m' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Confirm' }).click();

  await expect(page.getByLabel('Success')).toContainText('Your LPA has been saved.');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Return to task list' }).click();
  await page.getByRole('link', { name: 'Pay for the LPA' }).click();

  await expect(page.locator('h1')).toContainText('Paying for your LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Are you applying for any type of fee discount or exemption?');
  await page.getByRole('radio', { name: 'No' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();
  await page.getByRole('textbox', { name: 'Card number' }).fill('4444333322221111');
  await page.getByRole('textbox', { name: 'Month' }).fill('10');
  await page.getByRole('textbox', { name: 'Year' }).fill('27');
  await page.getByRole('textbox', { name: 'Name on card' }).fill('Card Name');
  await page.getByRole('textbox', { name: 'Card security code' }).fill('123');
  await page.getByRole('textbox', { name: 'Address line 1' }).fill('1 Richmond Place');
  await page.getByRole('textbox', { name: 'Postcode' }).fill('B14 7ED');
  await page.getByRole('textbox', { name: 'Town or city' }).fill('Birmingham');
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.getByLabel('Important')).toContainText('This is a test page. No money will be taken.');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Confirm payment' }).click();

  await expect(page.locator('h1')).toContainText('Payment received');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();
  await page.getByRole('link', { name: 'Confirm your identity' }).click();

  await expect(page.locator('h1')).toContainText('Confirm your identity');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.getByLabel('Success').locator('h2')).toContainText('You have successfully confirmed your identity');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Return to task list' }).click();
  await page.getByRole('link', { name: 'Sign the LPA' }).click();

  await expect(page.locator('h1')).toContainText('How to sign your LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Start' }).click();

  await expect(page.locator('h1')).toContainText('Your LPA will be registered in English');
  await page.getByRole('radio', { name: 'Register my LPA in English' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Save and continue' }).click();

  await expect(page.locator('h1')).toContainText('Read your LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Your legal rights and responsibilities');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue to signing page' }).click();

  await expect(page.locator('h1')).toContainText('Sign your LPA');
  await page.getByRole('checkbox', { name: 'I want to sign this LPA as a' }).check();
  await page.getByRole('checkbox', { name: 'I want to apply to register' }).check();
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Submit my signature' }).click();
  await page.getByText('GOV.UK One Login GOV.UK One').click();

  await expect(page.locator('h1')).toContainText('Witnessing your signature');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('Charlie Cooper, confirm you witnessed the donor sign their LPA');
  await page.getByRole('textbox', { name: 'Enter code' }).fill('1234');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('button', { name: 'Continue' }).click();

  await expect(page.locator('h1')).toContainText('You’ve submitted your LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Continue' }).click();
  await page.getByRole('link', { name: 'Check LPA progress' }).click();

  await expect(page.locator('#main-content')).toContainText('Check the progress of your LPA');
        await screenshot(page)
    await extractTextFromMainAndSave(page)
  await page.getByRole('link', { name: 'Return to dashboard' }).click();
  await page.getByRole('link', { name: 'View LPA' }).click();

  await expect(page.locator('h1')).toContainText('View your LPA');
});
