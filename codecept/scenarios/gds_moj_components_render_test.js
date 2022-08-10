Feature('GDS and MOJ components are available');

Scenario('displays a GDS summary element', ({ I}) => {
    I.amOnPage('/home')
    I.seeElement(
        locate('summary').withText('Help with nationality')
    )
    I.runAccessibilityChecks()
});

Scenario('displays a MOJ password reveal element', ({ I}) => {
    I.amOnPage('/home')
    I.seeInField('[data-module=moj-password-reveal]', '1234ABC!')
    I.seeAttributesOnElements('[data-module=moj-password-reveal]', {'type': 'password'})
    I.click('Show')
    I.seeAttributesOnElements('[data-module=moj-password-reveal]', {'type': 'text'})
    I.runAccessibilityChecks()
})
