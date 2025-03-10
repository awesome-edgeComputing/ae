#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import stat
import time
import click
from typing import Dict, Any


class FileUtil:
    """文件工具类，提供文件元信息获取等功能"""

    @staticmethod
    def get_file_metadata(file_path: str) -> Dict[str, Any]:
        """
        获取文件的元信息
        
        Args:
            file_path (str): 文件路径
            
        Returns:
            Dict[str, Any]: 包含文件元信息的字典，包括：
                - size: 文件大小（字节）
                - created_time: 创建时间
                - modified_time: 最后修改时间
                - accessed_time: 最后访问时间
                - mode: 文件权限模式
                - is_dir: 是否是目录
                - is_file: 是否是文件
                - is_symlink: 是否是符号链接
                - owner: 文件所有者ID
                - group: 文件所属组ID
        
        Raises:
            FileNotFoundError: 当文件不存在时抛出此异常
        """
        if not os.path.exists(file_path):
            raise FileNotFoundError(f"文件不存在：{file_path}")
        
        stat_info = os.stat(file_path)
        
        return {
            "size": stat_info.st_size,
            "created_time": time.ctime(stat_info.st_ctime),
            "modified_time": time.ctime(stat_info.st_mtime),
            "accessed_time": time.ctime(stat_info.st_atime),
            "mode": stat.filemode(stat_info.st_mode),
            "is_dir": os.path.isdir(file_path),
            "is_file": os.path.isfile(file_path),
            "is_symlink": os.path.islink(file_path),
            "owner": stat_info.st_uid,
            "group": stat_info.st_gid
        }

    @staticmethod
    def format_metadata(metadata: Dict[str, Any], format_type: str = "text") -> str:
        """
        格式化元信息为指定格式的字符串
        
        Args:
            metadata (Dict[str, Any]): 文件元信息字典
            format_type (str): 输出格式，支持 "text"、"json" 和 "yaml"
            
        Returns:
            str: 格式化后的字符串
        """
        if format_type == "json":
            import json
            return json.dumps(metadata, indent=2)
        elif format_type == "yaml":
            import yaml
            return yaml.dump(metadata, default_flow_style=False)
        else:
            return "\n".join([f"{key}: {value}" for key, value in metadata.items()])


@click.group()
def commands():
    """文件操作相关命令"""
    pass


@commands.command()
@click.argument('file_path', type=click.Path(exists=True))
@click.option("--format", type=click.Choice(['text', 'json', 'yaml']), default="text", help="输出格式")
def info(file_path, format):
    """获取文件的元信息"""
    try:
        file_util = FileUtil()
        metadata = file_util.get_file_metadata(file_path)
        print(file_util.format_metadata(metadata, format))
    except Exception as e:
        print(f"错误: {str(e)}")


if __name__ == "__main__":
    commands()
