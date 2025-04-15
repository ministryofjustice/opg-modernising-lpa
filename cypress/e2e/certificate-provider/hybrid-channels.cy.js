const { randomShareCode } = require("../../support/e2e");

describe('Hybrid certificate provider', () => {
    context('starts online, submits on paper, tries to access online again', () => {
        let sub = ''
        beforeEach(() => {
            sub = randomShareCode()
            cy.visit(`/fixtures?redirect=&progress=certificateProviderAccessCodeUsed&certificateProviderChannel=paper&certificateProviderSub=${sub}`);

            cy.visit('/certificate-provider-start')
            cy.contains('a', 'Start').click()

            cy.origin('http://localhost:7012', { args: { sub } }, ({ sub }) => {
                cy.get('input[name="subject"]').check('email', { force: true });
                cy.get('input[name="email"]').invoke('val', sub);
                cy.contains('button', 'Continue').click();
            });
            cy.url().should('contain', '/dashboard')
        });

        it('does not see LPA on dashboard', { pageLoadTimeout: 6000 }, () => {
            cy.contains('I’m a certificate provider').click()
            cy.contains('Property and affairs');

            cy.contains('a', 'Go to task list').click();

            cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
                .invoke('text')
                .then((uid) => {
                    cy.request({
                        method: 'POST',
                        url: 'http://localhost:9001/emit/opg.poas.sirius/certificate-provider-submission-completed',
                        body: {
                            uid: uid.trim(),
                        },
                    }).then((response) => {
                        expect(response.status).to.eq(200);

                        cy.visit('/dashboard')
                        cy.waitForTextVisibilityByReloading('main', 'I’m a certificate provider', false)
                    });
                });
        });
    });
});
