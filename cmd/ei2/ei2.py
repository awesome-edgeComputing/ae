#!/usr/bin/env python3

import os
import sys
import click

# Add the package directory to sys.path for PyInstaller
if getattr(sys, 'frozen', False):
    # Running in a PyInstaller bundle
    base_path = sys._MEIPASS
    package_path = os.path.join(base_path, 'ei2_cmd')
    if os.path.exists(package_path):
        sys.path.insert(0, os.path.dirname(package_path))

from ei2_cmd import FileUtil
from idevz import Idevz

@click.group()
def main():
    """EI2 - Extended Intelligence Interface"""
    pass

@main.command(name="f")
@click.argument('file_path', type=click.Path(exists=True))
@click.option("--format", type=str, default="json", help="输出格式: json, yaml, text")
def file_info(file_path, format):
    """获取文件的元信息"""
    try:
        file_util = FileUtil()
        metadata = file_util.get_file_metadata(file_path)
        print(file_util.format_metadata(metadata))
    except Exception as e:
        print(f"错误: {str(e)}")

@main.command()
def hello():
    """Print OK"""
    print("OK")

@main.command()
def idevz():
    """Print 周晶"""
    print("周晶")

@main.command(name="k")
def print_idevz():
    """Print 周晶 using Idevz class"""
    idevz = Idevz()
    idevz.print()

@main.command(name="y")
@click.option("--name", type=str, default="k-49")
def print_name(name):
    """Print k-49"""
    print(name)

if __name__ == "__main__":
    main() 