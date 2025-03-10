from setuptools import setup, find_packages

setup(
    name='ei2',
    version='0.1.0',
    packages=['ei2_cmd'],
    py_modules=['ei2'],
    install_requires=[
        'click>=7.1.2,<8.1.8',
        'requests>=2.25.0,<3.0.0',
        'tqdm>=4.60.0,<5.0.0',
        'numpy>=1.19.5,<2.0.0',
        'PyYAML>=5.4.1,<7.0.0',
        'onnx>=1.12.0,<2.0.0',
    ],
    entry_points={
        'console_scripts': [
            'ei2=ei2:main',
        ],
    },
) 