FROM public.ecr.aws/lambda/python:3.13

WORKDIR ${LAMBDA_TASK_ROOT}

COPY lambda/egress_checker ${LAMBDA_TASK_ROOT}

RUN pip install --no-cache-dir --requirement requirements.txt

CMD [ "lambda_handler" ]
